package marshal

import (
	"errors"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

func TestParseOrderedMap(t *testing.T) {
	g := ghost.New(t)

	index := 0
	ms := yaml.MapSlice{
		{Key: "foo", Value: "bar"},
		{Key: "bar", Value: "baz"},
	}

	defer func() { g.Should(be.Equal(2, index)) }()

	assign := func(name string, text []byte) error {
		defer func() { index++ }()

		key, ok := ms[index].Key.(string)
		g.Should(be.True(ok))

		g.Should(be.Equal(name, key))

		value, ok := ms[index].Value.(string)
		g.Should(be.True(ok))

		g.Should(be.Equal(value+"\n", string(text)))

		return nil
	}

	got, err := ParseOrderedMap(ms, assign)
	g.NoError(err)

	want := []string{"foo", "bar"}
	g.Should(be.DeepEqual(want, got))
}

func TestParseOrderedMap_stops_on_failure(t *testing.T) {
	g := ghost.New(t)

	index := 0
	ms := yaml.MapSlice{
		{Key: "foo", Value: "bar"},
		{Key: "bar", Value: "baz"},
	}

	defer func() { g.Should(be.Equal(1, index)) }()

	assign := func(name string, text []byte) error {
		index++
		return errors.New("uh oh")
	}

	_, err := ParseOrderedMap(ms, assign)
	g.Should(be.ErrorEqual("uh oh", err))
}

func TestParseOrderedMap_validates_key(t *testing.T) {
	g := ghost.New(t)

	ms := yaml.MapSlice{
		{Key: []string{"foo", "bar"}, Value: "bar"},
	}

	assign := func(name string, text []byte) error {
		return nil
	}

	_, err := ParseOrderedMap(ms, assign)
	g.Should(be.ErrorEqual(`["foo" "bar"] is not a valid key name`, err))
}
