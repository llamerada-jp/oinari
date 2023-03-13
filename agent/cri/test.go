package cri

import (
	"fmt"

	"github.com/stretchr/testify/suite"
)

type CriSuite struct {
	suite.Suite
}

func (suite *CriSuite) TestAa() {
	fmt.Println("test!!!")
}
