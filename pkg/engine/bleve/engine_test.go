package bleve_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/jsteenb2/search"
	"github.com/jsteenb2/search/pkg/engine/bleve"
	searchtest "github.com/jsteenb2/search/testing"
	"github.com/stretchr/testify/require"
)

func Test_Engine(t *testing.T) {
	initFn := func(t *testing.T) (search.Engine, string, func()) {
		tempDir := newTempDir(t)

		engine, err := bleve.NewEngine(bleve.IndexCfg{
			Name: "base",
			Path: path.Join(tempDir, "base.bleve"),
		})
		require.NoError(t, err)

		return engine, "base", func() {
			defer os.RemoveAll(tempDir)
		}
	}

	searchtest.TestSearchQueries(t, initFn)
}

func newTempDir(t *testing.T) string {
	t.Helper()

	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	return dir
}
