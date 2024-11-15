package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/db/pg"
)

type manager struct {
	db db.Transactor
}

// NewTransactionManager creates a new manager that meets the requirements of the interface db.TxManager
func NewTransactionManager(tr db.Transactor) db.TxManager {
	return &manager{
		db: tr,
	}
}

// transaction the main function that executes the user's handler in a transaction
func (m *manager) transaction(ctx context.Context, opts pgx.TxOptions, fn db.TxHandler) (err error) {
	// If it's a nested transaction, skip initialization of a new transaction and execute the handler
	tx, ok := ctx.Value(pg.TxKey).(pgx.Tx)
	if ok {
		return fn(ctx)
	}

	// Start a new transaction.
	tx, err = m.db.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "can't begin transaction")
	}

	// Put the transaction in the context.
	ctx = pg.MakeContextTx(ctx, tx)

	// Set up the defer function for the transaction to be rollbacked or commited.
	defer func() {
		// Recover from panic
		if r := recover(); r != nil {
			err = errors.Errorf("panic recovered: %v", r)
		}

		// Open a new transaction if an error occurred.
		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollback)
			}

			return
		}

		// If there were no errors, commit the transaction.
		err = tx.Commit(ctx)
		if err != nil {
			err = errors.Wrap(err, "tx commit failed")
		}
	}()

	// Execute a user's handler.
	// If the handler function returns an error, the transaction is rollbacked,
	// otherwise the transaction is commited.
	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	return err
}

func (m *manager) ReadCommitted(ctx context.Context, f db.TxHandler) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return m.transaction(ctx, txOpts, f)
}
