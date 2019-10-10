package disallow_test

import (
	"testing"

	"github.com/tvastar/disallow"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestDisallow(t *testing.T) {
	dir, cleanup, err := analysistest.WriteFiles(files)
	if err != nil {
		t.Fatal("Could not write files", err)
	}
	defer cleanup()

	names := disallow.Names(map[string]string{
		"panic":                              "Do not use panic. Handle/return errors or use log.Fatal",
		"context.WithValue":                  "Do not use context.WithValue",
		"net/http.StatusInternalServerError": "Do not, just do not.",
	})
	analysistest.Run(t, dir, names, "./...")
}

var files = map[string]string{
	"disallow.go": `
package disallow

import "context"
import "net/http"
import "fmt"

func mysampletest() {
     panic("yoyo") // want "Do not use panic. Handle/return errors or use log.Fatal"
}

func myCode(ctx context.Context) context.Context {
    return context.WithValue(ctx, "key", "value") // want "Do not use context.WithValue"
}

type Fake struct {}
func (f Fake) WithValue() bool {
     return false
}

func shouldNotTrigger(context Fake) bool {
     return context.WithValue()
}

func statusError() {
     fmt.Println("Should fail", http.StatusInternalServerError) // want "Do not, just do not."
}

`,
}
