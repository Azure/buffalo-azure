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
		desiredPath := path.Join(root, filename)
		if err := os.MkdirAll(path.Dir(desiredPath), os.ModePerm); err != nil {
			return err
		}
		if err := ioutil.WriteFile(desiredPath, contents, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// Clear removes all entries from a TemplateCache, so that they can be collected by the
// garbage collector.
func (c TemplateCache) Clear() {
	for k := range c {
		delete(c, k)
	}
}
