package main

import (
	"strconv"
	"strings"
	"testing"
	"time"

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
	rootCmd.ResetFlags()
	setupRootCmd()

	buf := new(strings.Builder)
	rootCmd.SetOutput(buf)
	s.output = buf
}

func (s *rootCmdSuite) run(args ...string) []string {
	rootCmd.SetArgs(args)
	s.Require().Nil(rootCmd.Execute())
	return strings.Split(strings.Trim(s.output.String(), "\n"), "\n")
}

func (s *rootCmdSuite) Test_Example() {
	assert := s.Assert()
	lines := s.run("example.munn")

	assert.NotEmpty(lines)
}

func (s *rootCmdSuite) Test_Example_Years() {
	assert := s.Assert()
	lines := s.run("example.munn", "--years", "100")

	if assert.NotEmpty(lines) {
		firstYear, _ := strconv.Atoi(strings.Split(lines[0], "-")[0])
		// Get the second-to-last line, since the final line will be the final balance
		lastYear, _ := strconv.Atoi(strings.Split(lines[len(lines)-2], "-")[0])
		assert.Equal(100, lastYear-firstYear, lines[0])
	}
}

func (s *rootCmdSuite) Test_Example_Retire() {
	assert := s.Assert()
	lines := s.run("example.munn", "--retire", "2080-01-01:25000")

	if assert.NotEmpty(lines) {
		lastLine := lines[len(lines)-1]
		assert.Equal("Retirement date: could not find", lastLine)
	}
}

func (s *rootCmdSuite) Test_Example_Retire_Years() {
	assert := s.Assert()
	lines := s.run("example.munn", "--retire", "2080-01-01:25000", "--years", "100")

	if assert.NotEmpty(lines) {
		firstYear, _ := strconv.Atoi(strings.Split(lines[0], "-")[0])
		// Get the thjrd-to-last line, since the last two lines are retirement info and final balance
		lastYear, _ := strconv.Atoi(strings.Split(lines[len(lines)-3], "-")[0])
		assert.Equal(100, lastYear-firstYear, lines[0])

		lastLine := strings.Split(lines[len(lines)-1], ": ")
		_, err := time.Parse("2006-01-02", lastLine[1])
		assert.Equal("Retirement date", lastLine[0])
		assert.Nil(err, "should have given a retirement date")
	}
}

func (s *rootCmdSuite) Test_Example_Stats() {
	assert := s.Assert()
	lines := s.run("example.munn", "--stats")

	if assert.True(len(lines) > 3, "should have at least 3 lines") {
		assert.Equal("Average monthly expenses", strings.Split(lines[0], ":")[0])
		assert.Equal("Average monthly income", strings.Split(lines[1], ":")[0])
		assert.Equal("Average monthly growth", strings.Split(lines[2], ":")[0])
	}
}
