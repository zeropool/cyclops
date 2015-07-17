package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkspace(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	assert.Equal("dockerfile", ws.Mode)
	assert.Equal("ubuntu:trusty", ws.Image)
	assert.Equal("ubuntu:trusty", ws.currentImage)
}

func TestSetImage(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	assert.Equal("ubuntu:trusty", ws.Image)
	assert.Equal("ubuntu:trusty", ws.currentImage)

	ws.SetImage("fedora")
	assert.Equal("fedora", ws.Image)
	assert.Equal("fedora", ws.currentImage)
}

func TestWorkspaceEval(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	res, err := ws.Eval("date", true)
	assert.NoError(err)
	assert.Equal("date", res.Command)
	assert.Equal("ubuntu:trusty", res.Image)
}

func TestWorkspaceSprint(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	out, err := ws.Sprint()
	assert.NoError(err)
	assert.Len(out, 1)
	assert.Equal("FROM ubuntu:trusty", out[0])
}

func TestWorkspaceWrite(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	os.Mkdir(".test", 0755)
	defer os.RemoveAll(".test")

	ws.SetImage("ubuntu:latest")
	ws.Run("touch /tmp")
	assert.NotPanics(func() {
		err := ws.Write(".test/Dockerfile")
		assert.NoError(err)
	})
}

// Workflow tests
func TestWorkflowRun(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	for i := 1; i < 4; i++ {
		res, err := ws.Run(fmt.Sprintf("cmd%v", i))
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("cmd%v", i), res.Command)
		assert.Equal(0, res.Code)
		assert.NotNil(res.Duration)
		assert.Equal(fmt.Sprintf("c%v", i), res.Id)
		assert.Equal("ubuntu:trusty", res.BaseImage)
		if i == 1 {
			assert.Equal("ubuntu:trusty", res.Image)
		} else {
			assert.Equal(fmt.Sprintf("i%v", (i-1)), res.Image)
		}
		assert.Equal(fmt.Sprintf("i%v", i), res.NewImage)

		assert.Equal(fmt.Sprintf("i%v", i), ws.currentImage)
	}
}

func TestWorkflowEval(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	for i := 1; i < 4; i++ {
		res, err := ws.Eval(fmt.Sprintf("cmd%v", i), true)
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("cmd%v", i), res.Command)
		assert.Equal(0, res.Code)
		assert.NotNil(res.Duration)
		assert.Equal(fmt.Sprintf("c%v", i), res.Id)
		assert.Equal("ubuntu:trusty", res.BaseImage)
		assert.Equal("ubuntu:trusty", res.Image)
		assert.Equal("", res.NewImage)

		assert.Equal("ubuntu:trusty", ws.currentImage)
	}
}

func TestWorkflowEvalCommit(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	for i := 1; i < 4; i++ {
		res, err := ws.Eval(fmt.Sprintf("cmd%v", i), true)
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("cmd%v", i), res.Command)
		assert.Equal(0, res.Code)
		assert.NotNil(res.Duration)
		assert.Equal(fmt.Sprintf("c%v", i), res.Id)
		assert.Equal("ubuntu:trusty", res.BaseImage)
		if i == 1 {
			assert.Equal("ubuntu:trusty", res.Image)
		} else {
			assert.Equal(fmt.Sprintf("i%v", (i-1)), res.Image)
		}
		assert.Equal("", res.NewImage)

		image, err := ws.CommitLast()
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("i%v", i), image)
	}

}

func TestWorkflowBack(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	run1, _ := ws.Run("cmd1")
	ws.Run("cmd2")
	ws.Run("cmd3")

	err := ws.back(2)
	assert.NoError(err)
	run4, _ := ws.Run("cmd4")

	expectedState := []string{"RUN cmd1", "RUN cmd4"}

	assert.Equal(run4.Image, run1.NewImage)
	assert.Equal(run4.NewImage, ws.currentImage)
	assert.True(ws.history[1].Deleted)
	assert.True(ws.history[2].Deleted)
	assert.Equal(expectedState, ws.state)
}
