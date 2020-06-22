/*
   Copyright 2020 Docker, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package framework

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	apicontext "github.com/docker/api/context"
	"github.com/docker/api/context/store"
)

// CliSuite is a helper struct that creates a configured context
// and captures the output of a command. it should be used in the
// same way as testify.suite.Suite
type CliSuite struct {
	suite.Suite
	ctx            context.Context
	writer         *os.File
	reader         *os.File
	OriginalStdout *os.File
	storeRoot      string
}

// BeforeTest is called by testify.suite
func (sut *CliSuite) BeforeTest(suiteName, testName string) {
	ctx := context.Background()
	ctx = apicontext.WithCurrentContext(ctx, "example")
	dir, err := ioutil.TempDir("", "store")
	require.Nil(sut.T(), err)
	s, err := store.New(
		store.WithRoot(dir),
	)
	require.Nil(sut.T(), err)

	err = s.Create("example", "example", "", store.ContextMetadata{})
	require.Nil(sut.T(), err)

	sut.storeRoot = dir

	ctx = store.WithContextStore(ctx, s)
	sut.ctx = ctx

	sut.OriginalStdout = os.Stdout
	r, w, err := os.Pipe()
	require.Nil(sut.T(), err)

	os.Stdout = w
	sut.writer = w
	sut.reader = r
}

// Context returns a configured context
func (sut *CliSuite) Context() context.Context {
	return sut.ctx
}

// GetStdOut returns the output of the command
func (sut *CliSuite) GetStdOut() string {
	err := sut.writer.Close()
	require.Nil(sut.T(), err)

	out, _ := ioutil.ReadAll(sut.reader)

	return string(out)
}

// AfterTest is called by testify.suite
func (sut *CliSuite) AfterTest(suiteName, testName string) {
	os.Stdout = sut.OriginalStdout
	err := os.RemoveAll(sut.storeRoot)
	require.Nil(sut.T(), err)
}