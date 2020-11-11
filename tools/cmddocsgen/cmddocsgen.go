// +build tools

/*
Copyright 2020 The kconnect Authors.

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

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fidelity/kconnect/internal/commands"
	"github.com/fidelity/kconnect/pkg/flags"
	_ "github.com/fidelity/kconnect/pkg/plugins" // Import all the plugins
	"github.com/fidelity/kconnect/pkg/provider"

	"github.com/spf13/cobra"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("you must supply a output folder")
	}
	outFolder := args[1]
	fmt.Printf("Generating cmdline docs. Output dir: %s\n", outFolder)

	rootCmd, _ := commands.RootCmd()
	err := genMarkdownTreeCustom(rootCmd, outFolder)
	if err != nil {
		log.Fatal(err)
	}
}

func genMarkdownTreeCustom(cmd *cobra.Command, dir string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genMarkdownTreeCustom(c, dir); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	basename = normalizeName(basename)
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Printf("Creating file: %s\n", filename)

	if err := genMarkdownCustom(cmd, f); err != nil {
		return err
	}
	return nil
}

func genMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	fmt.Printf("Printing docs for command: %s\n", cmd.Name())
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	isUseSubCmd := (cmd.Parent() != nil && cmd.Parent().Name() == "use")

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	buf.WriteString("## " + name + "\n\n")
	buf.WriteString(short + "\n\n")
	buf.WriteString("### Synopsis\n\n")
	buf.WriteString(long + "\n\n")

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}
	if isUseSubCmd {
		providerName := cmd.Name()
		if err := printIDPProtocolOptions(buf, cmd, name, providerName); err != nil {
			return err
		}
	}

	if hasSeeAlso(cmd) {
		buf.WriteString("### SEE ALSO\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			link := pname + ".md"
			link = strings.Replace(link, " ", "_", -1)
			link = normalizeName(link)
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", pname, link, parent.Short))
			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			link := cname + ".md"
			link = strings.Replace(link, " ", "_", -1)
			link = normalizeName(link)
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", cname, link, child.Short))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("\n")
	buf.WriteString("> NOTE: this page is auto-generated from the cobra commands\n")

	_, err := buf.WriteTo(w)
	return err
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Options\n\n```bash\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Options inherited from parent commands\n\n```bash\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}

func printIDPProtocolOptions(buf *bytes.Buffer, cmd *cobra.Command, name string, providerName string) error {
	buf.WriteString("### IDP Protocol Options\n\n")

	clusterProvider, err := provider.GetClusterProvider(providerName)
	if err != nil {
		return err
	}

	for _, idProviderName := range clusterProvider.SupportedIDs() {
		idProvider, err := provider.GetIdentityProvider(idProviderName)
		if err != nil {
			return nil
		}

		buf.WriteString(fmt.Sprintf("#### %s Options\n\n", strings.ToUpper(idProviderName)))
		buf.WriteString(fmt.Sprintf("Use `--idp-protocol=%s`\n\n", idProviderName))
		buf.WriteString("```bash\n")

		cfg, err := idProvider.ConfigurationItems(providerName)
		if err != nil {
			return err
		}

		fs, err := flags.CreateFlagsFromConfig(cfg)
		if err != nil {
			return err
		}
		fs.SetOutput(buf)
		fs.PrintDefaults()
		buf.WriteString("```\n\n")

	}

	return nil
}

func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		return true
	}
	return false
}

func normalizeName(name string) string {
	normalized := strings.Replace(name, "kconnect_", "", -1)
	if normalized == "kconnect.md" {
		normalized = "index.md"
	}

	return normalized
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
