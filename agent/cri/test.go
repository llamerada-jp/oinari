package cri

import (
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
	// expect the listRes empty
	listRes, err := suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Equal(0, len(listRes.Images))

	// expect there to be one image after pull a image
	pullRes, err := suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test1.container.json",
		},
	})
	test1ID := pullRes.ImageRef

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 1)
	suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test1.container.json", "go1.19", test1ID))

	// expect there to be two images after pull another image
	pullRes, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test2.container.json",
		},
	})
	test2ID := pullRes.ImageRef

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 2)

	if listRes.Images[0].Spec.Image == "http://localhost:8080/test1.container.json" {
		suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test1.container.json", "go1.19", test1ID))
		suite.True(checkImage(&listRes.Images[1], "http://localhost:8080/test2.container.json", "go1.18", test2ID))
	} else {
		suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test2.container.json", "go1.18", test2ID))
		suite.True(checkImage(&listRes.Images[1], "http://localhost:8080/test1.container.json", "go1.19", test1ID))
	}
	suite.NotEqual(listRes.Images[0].ID, listRes.Images[1].ID)

	// expect there to be two images after pull the same image
	pullRes, err = suite.cri.PullImage(&PullImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test1.container.json",
		},
	})
	suite.NoError(err)
	suite.Equal(test1ID, pullRes.ImageRef)

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 2)

	// expect there to be one image after remove the image
	_, err = suite.cri.RemoveImage(&RemoveImageRequest{
		Image: ImageSpec{
			Image: "http://localhost:8080/test1.container.json",
		},
	})
	suite.NoError(err)

	listRes, err = suite.cri.ListImages(&ListImagesRequest{})
	suite.NoError(err)
	suite.Len(listRes.Images, 1)
	suite.True(checkImage(&listRes.Images[0], "http://localhost:8080/test2.container.json", "go1.18", test2ID))
}

func checkImage(image *Image, url, runtime, id string) bool {
	if len(image.ID) == 0 || image.ID != id {
		return false
	}

	if len(image.Spec.Image) == 0 || image.Spec.Image != url {
		return false
	}

	if len(image.Runtime) == 0 || image.Runtime != runtime {
		return false
	}

	return true
}
