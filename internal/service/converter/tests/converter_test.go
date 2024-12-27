package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	converterMocks "github.com/chistyakoviv/converter/internal/converter/mocks"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/model"
	converterService "github.com/chistyakoviv/converter/internal/service/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConverterService(t *testing.T) {
	wd, err := os.Getwd()

	if err != nil {
		t.Fatalf("failed to get working directory: %s", err)
	}

	var (
		ctx          = context.Background()
		logger       = dummy.NewDummyLogger()
		configPath   = "config/local.yaml"
		defaultsPath = "config/defaults.yaml"
	)

	type testcase struct {
		name               string
		conversion         *model.Conversion
		err                string
		configPath         string
		defaultsPath       string
		mockImageConverter func(tc *testcase) *converterMocks.MockImageConverter
		mockVideoConverter func(tc *testcase) *converterMocks.MockVideoConverter
	}

	cases := []testcase{
		{
			name: "File does not exist",
			conversion: &model.Conversion{
				Fullpath: "/files/images/non-existent.jpg",
				Path:     "/files/images",
				Filestem: "non-existent",
				Ext:      "jpg",
				ConvertTo: []model.ConvertTo{
					{
						Ext: "webp",
					},
				},
			},
			err:          fmt.Sprintf("file '%s/files/images/non-existent.jpg' does not exist", wd),
			configPath:   configPath,
			defaultsPath: defaultsPath,
			mockImageConverter: func(tc *testcase) *converterMocks.MockImageConverter {
				mockImageConverter := converterMocks.NewMockImageConverter(t)
				return mockImageConverter
			},
			mockVideoConverter: func(tc *testcase) *converterMocks.MockVideoConverter {
				mockVideoConverter := converterMocks.NewMockVideoConverter(t)
				return mockVideoConverter
			},
		},
		{
			name: "Wrong file format",
			conversion: &model.Conversion{
				Fullpath: "/files/other/test.txt",
				Path:     "/files/other",
				Filestem: "test",
				Ext:      "txt",
				ConvertTo: []model.ConvertTo{
					{
						Ext: "webp",
					},
				},
			},
			err:          fmt.Sprintf("the file is not an image or video: %s/files/other/test.txt", wd),
			configPath:   configPath,
			defaultsPath: defaultsPath,
			mockImageConverter: func(tc *testcase) *converterMocks.MockImageConverter {
				mockImageConverter := converterMocks.NewMockImageConverter(t)
				return mockImageConverter
			},
			mockVideoConverter: func(tc *testcase) *converterMocks.MockVideoConverter {
				mockVideoConverter := converterMocks.NewMockVideoConverter(t)
				return mockVideoConverter
			},
		},
		{
			name: "Successfull image conversion",
			conversion: &model.Conversion{
				Fullpath: "/files/images/gen.jpg",
				Path:     "/files/images",
				Filestem: "gen",
				Ext:      "jpg",
				ConvertTo: []model.ConvertTo{
					{
						Ext: "webp",
					},
				},
			},
			configPath:   configPath,
			defaultsPath: defaultsPath,
			mockImageConverter: func(tc *testcase) *converterMocks.MockImageConverter {
				mockImageConverter := converterMocks.NewMockImageConverter(t)
				src, _ := tc.conversion.AbsoluteSourcePath()
				dest, _ := tc.conversion.AbsoluteDestinationPath(tc.conversion.ConvertTo[0])
				mockImageConverter.On(
					"Convert",
					src,
					dest,
					mock.Anything,
				).Return(nil).Once()
				return mockImageConverter
			},
			mockVideoConverter: func(tc *testcase) *converterMocks.MockVideoConverter {
				mockVideoConverter := converterMocks.NewMockVideoConverter(t)
				return mockVideoConverter
			},
		},
		{
			name: "Successfull video conversion",
			conversion: &model.Conversion{
				Fullpath: "/files/videos/gen.mp4",
				Path:     "/files/videos",
				Filestem: "gen",
				Ext:      "mp4",
				ConvertTo: []model.ConvertTo{
					{
						Ext: "webm",
					},
				},
			},
			configPath:   configPath,
			defaultsPath: defaultsPath,
			mockImageConverter: func(tc *testcase) *converterMocks.MockImageConverter {
				mockImageConverter := converterMocks.NewMockImageConverter(t)
				return mockImageConverter
			},
			mockVideoConverter: func(tc *testcase) *converterMocks.MockVideoConverter {
				mockVideoConverter := converterMocks.NewMockVideoConverter(t)
				src, _ := tc.conversion.AbsoluteSourcePath()
				dest, _ := tc.conversion.AbsoluteDestinationPath(tc.conversion.ConvertTo[0])
				mockVideoConverter.On(
					"Convert",
					src,
					dest,
					mock.Anything,
				).Return(nil).Once()
				return mockVideoConverter
			},
		},
		{
			name: "Merge video configuration of 2 formats with suffixes",
			conversion: &model.Conversion{
				Fullpath: "/files/videos/gen.mp4",
				Path:     "/files/videos",
				Filestem: "gen",
				Ext:      "mp4",
				ConvertTo: []model.ConvertTo{
					{
						Ext: "webm",
						Optional: map[string]interface{}{
							"replace_orig_ext": true,
							"suffix":           ".vp9",
						},
					},
					{
						Ext: "webm",
						Optional: map[string]interface{}{
							"replace_orig_ext": true,
							"suffix":           ".av1",
						},
						ConvConf: map[string]interface{}{
							"crf": "50",
						},
					},
				},
			},
			configPath:   configPath,
			defaultsPath: "config/defaults_suffixed.yaml",
			mockImageConverter: func(tc *testcase) *converterMocks.MockImageConverter {
				mockImageConverter := converterMocks.NewMockImageConverter(t)
				return mockImageConverter
			},
			mockVideoConverter: func(tc *testcase) *converterMocks.MockVideoConverter {
				mockVideoConverter := converterMocks.NewMockVideoConverter(t)
				src, _ := tc.conversion.AbsoluteSourcePath()
				destVP9, _ := tc.conversion.AbsoluteDestinationPath(tc.conversion.ConvertTo[0])
				destAV1, _ := tc.conversion.AbsoluteDestinationPath(tc.conversion.ConvertTo[1])
				mockVideoConverter.On(
					"Convert",
					src,
					destVP9,
					converter.ConversionConfig{
						"c:v": "libvpx-vp9",
						"c:a": "libopus",
						"crf": "35",
					},
				).Return(nil).Once()
				mockVideoConverter.On(
					"Convert",
					src,
					destAV1,
					converter.ConversionConfig{
						"c:v": "libaom-av1",
						"c:a": "libopus",
						"crf": "50",
					},
				).Return(nil).Once()
				return mockVideoConverter
			},
		},
		{
			name: "Merge video configuration of 2 formats with and without suffix",
			conversion: &model.Conversion{
				Fullpath: "/files/videos/gen.mp4",
				Path:     "/files/videos",
				Filestem: "gen",
				Ext:      "mp4",
				ConvertTo: []model.ConvertTo{
					{
						Ext: "webm",
						Optional: map[string]interface{}{
							"replace_orig_ext": true,
						},
					},
					{
						Ext: "webm",
						Optional: map[string]interface{}{
							"replace_orig_ext": true,
							"suffix":           ".av1",
						},
						ConvConf: map[string]interface{}{
							"crf": "50",
						},
					},
				},
			},
			configPath:   configPath,
			defaultsPath: "config/defaults_mixed_suffixes.yaml",
			mockImageConverter: func(tc *testcase) *converterMocks.MockImageConverter {
				mockImageConverter := converterMocks.NewMockImageConverter(t)
				return mockImageConverter
			},
			mockVideoConverter: func(tc *testcase) *converterMocks.MockVideoConverter {
				mockVideoConverter := converterMocks.NewMockVideoConverter(t)
				src, _ := tc.conversion.AbsoluteSourcePath()
				destVP9, _ := tc.conversion.AbsoluteDestinationPath(tc.conversion.ConvertTo[0])
				destAV1, _ := tc.conversion.AbsoluteDestinationPath(tc.conversion.ConvertTo[1])
				mockVideoConverter.On(
					"Convert",
					src,
					destVP9,
					converter.ConversionConfig{
						"c:v": "libvpx-vp9",
						"c:a": "libopus",
						"crf": "35",
					},
				).Return(nil).Once()
				mockVideoConverter.On(
					"Convert",
					src,
					destAV1,
					converter.ConversionConfig{
						"c:v": "libaom-av1",
						"c:a": "libopus",
						"crf": "50",
					},
				).Return(nil).Once()
				return mockVideoConverter
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockImageConverter := tc.mockImageConverter(&tc)
			mockVideoConverter := tc.mockVideoConverter(&tc)

			serv, _ := converterService.NewService(
				config.MustLoad(&config.ConfigOptions{
					ConfigPath:   tc.configPath,
					DefaultsPath: tc.defaultsPath,
				}),
				logger,
				mockImageConverter,
				mockVideoConverter,
			)

			err := serv.Convert(ctx, tc.conversion)
			if tc.err != "" {
				assert.Equal(t, err.Error(), tc.err)
			} else {
				assert.NoError(t, err)
			}

			mockImageConverter.AssertExpectations(t)
			mockVideoConverter.AssertExpectations(t)
		})
	}
}
