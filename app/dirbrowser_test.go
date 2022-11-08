package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTmplData_Breadcrumbs(t *testing.T) {
	check := func(url string, bc []breadcrumb) {
		t.Helper()
		td := tmplData{URL: url}
		assert.Equal(t, bc, td.Breadcrumbs())
	}

	check("/", nil)
	check("/foo", []breadcrumb{{Name: "foo", Href: "/foo"}})
	check("/foo/bar", []breadcrumb{{Name: "foo", Href: "/foo"}, {Name: "bar", Href: "/foo/bar"}})
	check("/foo/bar/", []breadcrumb{{Name: "foo", Href: "/foo"}, {Name: "bar", Href: "/foo/bar"}})
}
