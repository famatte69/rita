package commands

import (
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/ocmdev/rita/database"
	"github.com/ocmdev/rita/datatypes/urls"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func init() {
	longURLs := cli.Command{

		Name:  "show-long-urls",
		Usage: "Print the longest urls",
		Flags: []cli.Flag{
			humanFlag,
			databaseFlag,
			configFlag,
		},
		Action: func(c *cli.Context) error {
			if c.String("database") == "" {
				return cli.NewExitError("Specify a database with -d", -1)
			}

			res := database.InitResources(c.String("config"))

			var urls []urls.URL
			coll := res.DB.Session.DB(c.String("database")).C(res.Config.T.Urls.UrlsTable)

			coll.Find(nil).Sort("-length").All(&urls)

			if len(urls) == 0 {
				return cli.NewExitError("No results were found for "+c.String("database"), -1)
			}

			if c.Bool("human-readable") {
				err := showURLsHuman(urls)
				if err != nil {
					return cli.NewExitError(err.Error(), -1)
				}
			}
			err := showURLs(urls)
			if err != nil {
				return cli.NewExitError(err.Error(), -1)
			}
			return nil
		},
	}
	vistedURLs := cli.Command{

		Name:  "show-most-visited-urls",
		Usage: "Print the most visited urls",
		Flags: []cli.Flag{
			humanFlag,
			databaseFlag,
		},
		Action: func(c *cli.Context) error {
			if c.String("database") == "" {
				return cli.NewExitError("Specify a database with -d", -1)
			}

			res := database.InitResources("")

			var urls []urls.URL
			coll := res.DB.Session.DB(c.String("database")).C(res.Config.T.Urls.UrlsTable)

			coll.Find(nil).Sort("-count").All(&urls)

			if len(urls) == 0 {
				return cli.NewExitError("No results were found for "+c.String("database"), -1)
			}

			if c.Bool("human-readable") {
				err := showURLsHuman(urls)
				if err != nil {
					return cli.NewExitError(err.Error(), -1)
				}
			}
			err := showURLs(urls)
			if err != nil {
				return cli.NewExitError(err.Error(), -1)
			}
			return nil
		},
	}
	bootstrapCommands(longURLs, vistedURLs)
}

func showURLs(urls []urls.URL) error {
	tmpl := "{{.URL}},{{.URI}},{{.Length}},{{.Count}},{{range $idx, $ip := .IPs}}{{if $idx}} {{end}}{{ $ip }}{{end}}\n"

	out, err := template.New("urls").Parse(tmpl)
	if err != nil {
		return err
	}

	for _, url := range urls {
		err := out.Execute(os.Stdout, url)
		if err != nil {
			fmt.Fprintf(os.Stdout, "ERROR: Template failure: %s\n", err.Error())
		}
	}
	return nil
}

func showURLsHuman(urls []urls.URL) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(50)
	table.SetHeader([]string{"URL", "URI", "Length", "Times Visted"})
	for _, url := range urls {
		if len(url.URL) > 50 {
			url.URL = url.URL[0:47] + "..."
		}
		if len(url.URI) > 50 {
			url.URI = url.URI[0:47] + "..."
		}
		table.Append([]string{
			url.URL, url.URI, strconv.FormatInt(url.Length, 10), strconv.FormatInt(url.Count, 10),
		})
	}
	table.Render()
	return nil
}