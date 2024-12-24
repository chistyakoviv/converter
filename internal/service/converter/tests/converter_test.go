package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/chistyakoviv/converter/internal/config"
	converterMocks "github.com/chistyakoviv/converter/internal/converter/mocks"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/service/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConverterService(t *testing.T) {
	t.Skip("Skipping converter service test")
	wd, err := os.Getwd()

	if err != nil {
		t.Fatalf("failed to get working directory: %s", err)
	}

	var (
		ctx    = context.Background()
		logger = dummy.NewDummyLogger()
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
			err: fmt.Sprintf("file '%s/files/images/non-existent.jpg' does not exist", wd),
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
			err: fmt.Sprintf("the file is not an image or video: %s/files/other/test.txt", wd),
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockImageConverter := tc.mockImageConverter(&tc)
			mockVideoConverter := tc.mockVideoConverter(&tc)

			serv, _ := converter.NewService(
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
