package main

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type rootCmdSuite struct {
	suite.Suite

	output *strings.Builder
}

func Test_RootCmd(t *testing.T) {
	suite.Run(t, &rootCmdSuite{})
}

func (s *rootCmdSuite) SetupTest() {
	buf := new(strings.Builder)
	rootCmd.ResetFlags()
	setupRootCmd()
	rootCmd.SetOut(buf)
	s.output = buf
}

func (s *rootCmdSuite) run(args ...string) {
	rootCmd.SetArgs(args)
	require := s.Require()
	require.Nil(rootCmd.ParseFlags(args))
	require.Nil(rootCmd.Execute())
}

func (s *rootCmdSuite) lines() []string {
	return strings.Split(strings.Trim(s.output.String(), "\n"), "\n")
}

func (s *rootCmdSuite) Test_Example() {
	s.run("example.munn")
	assert := s.Assert()
	lines := s.lines()

	assert.NotEmpty(lines)
}

func (s *rootCmdSuite) Test_Example_Years() {
	s.run("example.munn", "--years", "100")
	assert := s.Assert()
	lines := s.lines()

	if assert.NotEmpty(lines) {
		firstYear, _ := strconv.Atoi(strings.Split(lines[0], "-")[0])
		lastYear, _ := strconv.Atoi(strings.Split(lines[len(lines)-1], "-")[0])
		assert.Equal(100, lastYear-firstYear, lines[0])
	}
}

func (s *rootCmdSuite) Test_Example_Retire() {
	s.run("example.munn", "--retire", "2080-01-01:25000")
	assert := s.Assert()
	lines := s.lines()

	if assert.NotEmpty(lines) {
		lastLine := lines[len(lines)-1]
		assert.Equal("Retirement date: could not find", lastLine)
	}
}

func (s *rootCmdSuite) Test_Example_Stats() {
	s.run("example.munn", "--stats")
	assert := s.Assert()
	lines := s.lines()

	if assert.True(len(lines) > 3, "should have at least 3 lines") {
		assert.Equal("Average monthly expenses", strings.Split(lines[0], ":")[0])
		assert.Equal("Average monthly income", strings.Split(lines[1], ":")[0])
		assert.Equal("Average monthly growth", strings.Split(lines[2], ":")[0])
	}
}
