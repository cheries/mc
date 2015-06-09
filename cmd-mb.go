/*
 * Minio Client (C) 2014, 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"errors"
	"fmt"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/client"
	"github.com/minio/mc/pkg/console"
	"github.com/minio/minio/pkg/iodine"
)

// Help message.
var mbCmd = cli.Command{
	Name:   "mb",
	Usage:  "Make a bucket or folder",
	Action: runMakeBucketCmd,
	CustomHelpTemplate: `NAME:
   mc {{.Name}} - {{.Usage}}

USAGE:
   mc {{.Name}} TARGET [TARGET...] {{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .Flags}}

FLAGS:
   {{range .Flags}}{{.}}
   {{end}}{{ end }}

EXAMPLES:
   1. Create a bucket on Amazon S3 object storage.
      $ mc {{.Name}} https://s3.amazonaws.com/public-document-store

   3. Make a directory on local filesystem, including its parent directories as needed.
      $ mc {{.Name}} ~/

   3. Create a bucket on Minio object storage.
      $ mc {{.Name}} https://play.minio.io:9000/mongodb-backup
`,
}

// runMakeBucketCmd is the handler for mc mb command
func runMakeBucketCmd(ctx *cli.Context) {
	if !ctx.Args().Present() || ctx.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(ctx, "mb", 1) // last argument is exit code
	}
	if !isMcConfigExist() {
		console.Fatals(ErrorMessage{
			Message: "Please run \"mc config generate\"",
			Error:   errors.New("\"mc\" is not configured"),
		})
	}
	config, err := getMcConfig()
	if err != nil {
		console.Fatals(ErrorMessage{
			Message: "Unable to read config file ‘" + mustGetMcConfigPath() + "’",
			Error:   err,
		})
	}
	targetURLConfigMap := make(map[string]*hostConfig)
	for _, arg := range ctx.Args() {
		targetURL, err := getExpandedURL(arg, config.Aliases)
		if err != nil {
			switch e := iodine.ToError(err).(type) {
			case errUnsupportedScheme:
				console.Fatals(ErrorMessage{
					Message: fmt.Sprintf("Unknown type of URL ‘%s’", e.url),
					Error:   e,
				})
			default:
				console.Fatals(ErrorMessage{
					Message: fmt.Sprintf("Unable to parse argument ‘%s’", arg),
					Error:   err,
				})
			}
		}
		targetConfig, err := getHostConfig(targetURL)
		if err != nil {
			console.Fatals(ErrorMessage{
				Message: fmt.Sprintf("Unable to read host configuration for ‘%s’ from config file ‘%s’", targetURL, mustGetMcConfigPath()),
				Error:   err,
			})
		}
		targetURLConfigMap[targetURL] = targetConfig
	}
	for targetURL, targetConfig := range targetURLConfigMap {
		errorMsg, err := doMakeBucketCmd(targetURL, targetConfig)
		if err != nil {
			console.Errors(ErrorMessage{
				Message: errorMsg,
				Error:   err,
			})
		}
	}
}

// doMakeBucketCmd -
func doMakeBucketCmd(targetURL string, targetConfig *hostConfig) (string, error) {
	var err error
	var clnt client.Client
	clnt, err = getNewClient(targetURL, targetConfig)
	if err != nil {
		msg := fmt.Sprintf("Unable to initialize client for ‘%s’", targetURL)
		return msg, iodine.New(err, nil)
	}
	return doMakeBucket(clnt, targetURL)
}

// doMakeBucket - wrapper around MakeBucket() API
func doMakeBucket(clnt client.Client, targetURL string) (string, error) {
	err := clnt.MakeBucket()
	if err != nil {
		msg := fmt.Sprintf("Failed to create bucket for URL ‘%s’", targetURL)
		return msg, iodine.New(err, nil)
	}
	return "", nil
}
