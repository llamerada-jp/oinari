package cri

import (
	"time"

	"github.com/llamerada-jp/oinari/agent/crosslink"
	"github.com/stretchr/testify/suite"
)

type CriSuite struct {
	suite.Suite
	cri CRI
}

func (suite *CriSuite) SetupSuite() {
	rootMpx := crosslink.NewMultiPlexer()
	cl := crosslink.NewCrosslink("crosslink", rootMpx)

	suite.cri = NewCRI(cl)
}

func (suite *CriSuite) TestImage() {
	// expect the images empty
	images, err := suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Equal(0, len(images.Images))

	// pull sample image
	res, err := suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/sample.container.json",
		},
	})
	suite.NotEqual(0, len(res.ImageRef))

	// expect there to be one image
	suite.Eventually(func() bool {
		images, err = suite.cri.ListImages(&ListImagesRequest{})
		suite.NoError(err)

		return len(images.Images) == 1
	}, 30*time.Second, 10*time.Second)

	suite.NotEqual(0, len(images.Images[0].ID))
	suite.Equal("go1.19", images.Images[0].Runtime)
	suite.Equal("http://localhost:8080/sample.container.json", images.Images[0].Spec.Image)

	// pull one more image
	res, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/sample.container.json",
		},
	})
	suite.NotEqual(0, len(res.ImageRef))

	// expect there to be two images
	suite.Eventually(func() bool {
		images, err = suite.cri.ListImages(&ListImagesRequest{})
		suite.NoError(err)
		return len(images.Images) == 2
	}, 30*time.Second, 10*time.Second)
}
