// Copyright 2015 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/tsuru/config"
	"github.com/tsuru/tsuru/app"
	"github.com/tsuru/tsuru/db"
	"github.com/tsuru/tsuru/db/dbtest"
	"github.com/tsuru/tsuru/io"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/provisiontest"
	"gopkg.in/check.v1"
)

type PlatformSuite struct{}

var _ = check.Suite(&PlatformSuite{})

func (s *PlatformSuite) SetUpTest(c *check.C) {
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "tsuru_api_platform_test")
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	dbtest.ClearAllCollections(conn.Apps().Database)
}

func (p *PlatformSuite) TestPlatformAdd(c *check.C) {
	provisioner := provisiontest.ExtensibleFakeProvisioner{
		FakeProvisioner: provisiontest.NewFakeProvisioner(),
	}
	oldProvisioner := app.Provisioner
	app.Provisioner = &provisioner
	defer func() {
		app.Provisioner = oldProvisioner
	}()
	var buf bytes.Buffer
	dockerfileURL := "http://localhost/Dockerfile"
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "test")
	writer.WriteField("dockerfile", dockerfileURL)
	fileWriter, err := writer.CreateFormFile("dockerfile_content", "Dockerfile")
	c.Assert(err, check.IsNil)
	fileWriter.Write([]byte("FROM tsuru/java"))
	writer.Close()
	request, _ := http.NewRequest("POST", "/platforms/add", &buf)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	result := platformAdd(recorder, request, nil)
	c.Assert(result, check.IsNil)
	c.Assert(recorder.Body.String(), check.Equals, "\nOK!\n")
}

func (p *PlatformSuite) TestPlatformUpdate(c *check.C) {
	provisioner := provisiontest.ExtensibleFakeProvisioner{
		FakeProvisioner: provisiontest.NewFakeProvisioner(),
	}
	oldProvisioner := app.Provisioner
	app.Provisioner = &provisioner
	defer func() {
		app.Provisioner = oldProvisioner
	}()
	err := app.PlatformAdd(provision.PlatformOptions{Name: "wat", Args: nil, Output: nil})
	c.Assert(err, check.IsNil)
	var buf bytes.Buffer
	dockerfileURL := "http://localhost/Dockerfile"
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dockerfile", dockerfileURL)
	writer.Close()
	request, _ := http.NewRequest("PUT", "/platforms/wat?:name=wat", &buf)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	result := platformUpdate(recorder, request, nil)
	c.Assert(result, check.IsNil)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var msg io.SimpleJsonMessage
	json.Unmarshal(b, &msg)
	c.Assert(errors.New(msg.Error), check.ErrorMatches, "")
}

func (p *PlatformSuite) TestPlatformUpdateOnlyDisableTrue(c *check.C) {
	provisioner := provisiontest.ExtensibleFakeProvisioner{
		FakeProvisioner: provisiontest.NewFakeProvisioner(),
	}
	oldProvisioner := app.Provisioner
	app.Provisioner = &provisioner
	defer func() {
		app.Provisioner = oldProvisioner
	}()
	err := app.PlatformAdd(provision.PlatformOptions{Name: "wat", Args: nil, Output: nil})
	c.Assert(err, check.IsNil)
	dockerfile_url := ""
	body := fmt.Sprintf("dockerfile=%s", dockerfile_url)
	request, _ := http.NewRequest("PUT", "/platforms/wat?:name=wat&disabled=true", strings.NewReader(body))
	request.Header.Add("Content-Type", "multipart/form-data")
	recorder := httptest.NewRecorder()
	result := platformUpdate(recorder, request, nil)
	c.Assert(result, check.IsNil)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var msg io.SimpleJsonMessage
	json.Unmarshal(b, &msg)
	c.Assert(errors.New(msg.Error), check.ErrorMatches, "")
}

func (p *PlatformSuite) TestPlatformUpdateDisableTrueAndDockerfile(c *check.C) {
	provisioner := provisiontest.ExtensibleFakeProvisioner{
		FakeProvisioner: provisiontest.NewFakeProvisioner(),
	}
	oldProvisioner := app.Provisioner
	app.Provisioner = &provisioner
	defer func() {
		app.Provisioner = oldProvisioner
	}()
	err := app.PlatformAdd(provision.PlatformOptions{Name: "wat", Args: nil, Output: nil})
	c.Assert(err, check.IsNil)
	dockerfile_url := "http://localhost/Dockerfile"
	body := fmt.Sprintf("dockerfile=%s", dockerfile_url)
	request, _ := http.NewRequest("PUT", "/platforms/wat?:name=wat&disabled=true", strings.NewReader(body))
	request.Header.Add("Content-Type", "multipart/form-data")
	recorder := httptest.NewRecorder()
	result := platformUpdate(recorder, request, nil)
	c.Assert(result, check.IsNil)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var msg io.SimpleJsonMessage
	json.Unmarshal(b, &msg)
	c.Assert(errors.New(msg.Error), check.ErrorMatches, "")
}

func (p *PlatformSuite) TestPlatformUpdateOnlyDisableFalse(c *check.C) {
	provisioner := provisiontest.ExtensibleFakeProvisioner{
		FakeProvisioner: provisiontest.NewFakeProvisioner(),
	}
	oldProvisioner := app.Provisioner
	app.Provisioner = &provisioner
	defer func() {
		app.Provisioner = oldProvisioner
	}()
	err := app.PlatformAdd(provision.PlatformOptions{Name: "wat", Args: nil, Output: nil})
	c.Assert(err, check.IsNil)
	dockerfile_url := ""
	body := fmt.Sprintf("dockerfile=%s", dockerfile_url)
	request, _ := http.NewRequest("PUT", "/platforms/wat?:name=wat&disabled=false", strings.NewReader(body))
	request.Header.Add("Content-Type", "multipart/form-data")
	recorder := httptest.NewRecorder()
	result := platformUpdate(recorder, request, nil)
	c.Assert(result, check.IsNil)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var msg io.SimpleJsonMessage
	json.Unmarshal(b, &msg)
	c.Assert(errors.New(msg.Error), check.ErrorMatches, "")
}

func (p *PlatformSuite) TestPlatformUpdateDisableFalseAndDockerfile(c *check.C) {
	provisioner := provisiontest.ExtensibleFakeProvisioner{
		FakeProvisioner: provisiontest.NewFakeProvisioner(),
	}
	oldProvisioner := app.Provisioner
	app.Provisioner = &provisioner
	defer func() {
		app.Provisioner = oldProvisioner
	}()
	err := app.PlatformAdd(provision.PlatformOptions{Name: "wat", Args: nil, Output: nil})
	c.Assert(err, check.IsNil)
	dockerfile_url := "http://localhost/Dockerfile"
	body := fmt.Sprintf("dockerfile=%s", dockerfile_url)
	request, _ := http.NewRequest("PUT", "/platforms/wat?:name=wat&disabled=false", strings.NewReader(body))
	request.Header.Add("Content-Type", "multipart/form-data")
	recorder := httptest.NewRecorder()
	result := platformUpdate(recorder, request, nil)
	c.Assert(result, check.IsNil)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var msg io.SimpleJsonMessage
	json.Unmarshal(b, &msg)
	c.Assert(errors.New(msg.Error), check.ErrorMatches, "")
}

func (p *PlatformSuite) TestPlatformRemove(c *check.C) {
	provisioner := provisiontest.ExtensibleFakeProvisioner{
		FakeProvisioner: provisiontest.NewFakeProvisioner(),
	}
	oldProvisioner := app.Provisioner
	app.Provisioner = &provisioner
	defer func() {
		app.Provisioner = oldProvisioner
	}()
	err := app.PlatformAdd(provision.PlatformOptions{Name: "test", Args: nil, Output: nil})
	c.Assert(err, check.IsNil)
	request, _ := http.NewRequest("DELETE", "/platforms/test?:name=test", nil)
	recorder := httptest.NewRecorder()
	err = platformRemove(recorder, request, nil)
	c.Assert(err, check.IsNil)
}
