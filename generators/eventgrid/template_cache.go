package eventgrid

import (
	"io/ioutil"
	"os"
	"path"
)

// TemplateCache stores files used as templates, so that they will always be distributed along
// with thie buffalo-azure binary.
type TemplateCache map[string][]byte

// Rehydrate writes the contents of each template file back to disk, rooted at the directory
// specified.
func (c TemplateCache) Rehydrate(root string) error {
	for filename, contents := range c {
		if err := ioutil.WriteFile(path.Join(root, filename), contents, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
