package tests

import (
	"os"
	"testing"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/converter/govips"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageConverter(t *testing.T) {
	var (
		logger         = dummy.NewDummyLogger()
		filesDir       = "files/images"
		filesOutputDir = filesDir + "/output"
		env            = config.EnvLocal
		imageConf      = config.Image{
			Threads: 4,
		}
	)

	// #nosec G301 -- this is test code and wide permissions are intentional
	outputDirErr := os.MkdirAll(filesOutputDir, 0777)
	require.NoError(t, outputDirErr)

	t.Cleanup(func() {
		err := os.RemoveAll(filesOutputDir)
		require.NoError(t, err, "Failed to remove images output dir")
	})

	type testcase struct {
		name string
		from string
		to   string
		err  string
		conf converter.ConversionConfig
	}

	cases := []testcase{
		{
			name: "Convert jpg to jpg",
			from: filesDir + "/gen.jpg",
			to:   filesOutputDir + "/gen-jpg-to-jpg.jpg",
			conf: nil,
		},
		{
			name: "Convert jpg to png",
			from: filesDir + "/gen.jpg",
			to:   filesOutputDir + "/gen-jpg-to-png.png",
			conf: nil,
		},
		{
			name: "Convert jpg to webp",
			from: filesDir + "/gen.jpg",
			to:   filesOutputDir + "/gen-jpg-to-webp.webp",
			conf: nil,
		},
		{
			name: "Convert jpg to avif",
			from: filesDir + "/gen.jpg",
			to:   filesOutputDir + "/gen-jpg-to-avif.avif",
			conf: nil,
		},
		{
			name: "Convert png to png",
			from: filesDir + "/gen.png",
			to:   filesOutputDir + "/gen-png-to-png.png",
			conf: nil,
		},
		{
			name: "Convert png to jpg",
			from: filesDir + "/gen.png",
			to:   filesOutputDir + "/gen-png-to-jpg.jpg",
			conf: nil,
		},
		{
			name: "Convert png to webp",
			from: filesDir + "/gen.png",
			to:   filesOutputDir + "/gen-png-to-webp.webp",
			conf: nil,
		},
		{
			name: "Convert png to avif",
			from: filesDir + "/gen.png",
			to:   filesOutputDir + "/gen-png-to-avif.avif",
			conf: nil,
		},
		{
			name: "Unsupported format",
			from: filesDir + "/gen.jpg",
			to:   filesOutputDir + "/gen-png-to-unsupported.ext",
			err:  "govips: unsupported format: ext",
			conf: nil,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				Env:   env,
				Image: imageConf,
			}

			t.Cleanup(func() {
				if err := os.Remove(tc.to); err != nil && !os.IsNotExist(err) {
					require.NoError(t, err, "Failed to remove generated file")
				}
			})

			converter := govips.NewImageConverter(cfg, logger)

			err := converter.Convert(tc.from, tc.to, tc.conf)
			if tc.err != "" {
				assert.Equal(t, err.Error(), tc.err)
				return
			}
			require.NoError(t, err)

			_, err = os.Stat(tc.to)
			assert.NoError(t, err, "File %s should exist", tc.to)
		})
	}
}
