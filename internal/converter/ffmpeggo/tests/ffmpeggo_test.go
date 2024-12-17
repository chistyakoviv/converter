package tests

import (
	"os"
	"testing"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/converter/ffmpeggo"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageConverter(t *testing.T) {
	var (
		logger         = dummy.NewDummyLogger()
		filesDir       = "files/videos"
		filesOutputDir = filesDir + "/output"
		env            = config.EnvLocal
		videoConf      = config.Video{
			Threads: 4,
		}
	)

	outputDirErr := os.MkdirAll(filesOutputDir, 0777)
	require.NoError(t, outputDirErr)

	t.Cleanup(func() {
		err := os.RemoveAll(filesOutputDir)
		require.NoError(t, err, "Failed to remove videos output dir")
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
			name: "Convert mp4 to mp4 with libx264 codec",
			from: filesDir + "/gen.mp4",
			to:   filesOutputDir + "/gen-mp4-to-mp4_h264.mp4",
			conf: converter.ConversionConfig{
				"c:v": "libx264",
				"crf": "40",
			},
		},
		{
			name: "Convert mp4 to webm with libvpx-vp9 codec",
			from: filesDir + "/gen.mp4",
			to:   filesOutputDir + "/gen-mp4-to-webm_vp9.mp4",
			conf: converter.ConversionConfig{
				"c:v": "libvpx-vp9",
				"crf": "40",
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				Env:   env,
				Video: videoConf,
			}

			t.Cleanup(func() {
				if err := os.Remove(tc.to); err != nil && !os.IsNotExist(err) {
					require.NoError(t, err, "Failed to remove generated file")
				}
			})

			converter := ffmpeggo.NewVideoConverter(cfg, logger)

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
