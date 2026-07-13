package eval

import (
	"sort"
	"strings"

	"src.elv.sh/pkg/eval/vals"
)

// ShellContext describes the live runtime state of the shell at a given moment.
// It is a generic, consumer-agnostic snapshot of the information that external
// tools (completion engines, prompt generators, etc.) typically need.
//
// All slices are sorted lexicographically and contain no duplicates.
type ShellContext struct {
	// Shell is the name of the shell, always "elvish".
	Shell string
	// Aliases are abbreviation keys from edit:abbr.
	Aliases []string
	// Builtins are special forms (if, while, ...) and builtin functions
	// (put, echo, ...).
	Builtins []string
	// Functions are user-defined function names (fn~ variables) from the
	// global scope, without the ~ suffix.
	Functions []string
	// Jobs are currently running background job identifiers. Elvish does not
	// maintain a job table, so this is always empty.
	Jobs []string
	// Variables are all variable names visible in the global scope plus the
	// builtin namespace, without sigils or suffixes.
	Variables []string
}

// Context returns a snapshot of the shell's live runtime state. It is safe to
// call concurrently with code execution; the returned value reflects the state
// at the moment of the call.
func (ev *Evaler) Context() ShellContext {
	ctx := ShellContext{Shell: "elvish"}

	ctx.Aliases = ev.aliasKeys()
	ctx.Builtins = ev.builtinNames()
	ctx.Functions = ev.functionNames()
	ctx.Variables = ev.variableNames()

	sort.Strings(ctx.Aliases)
	sort.Strings(ctx.Builtins)
	sort.Strings(ctx.Functions)
	sort.Strings(ctx.Variables)

	return ctx
}

// aliasKeys returns the keys of the edit:abbr map. The edit: namespace is
// added to the builtin namespace at editor initialization, so it may not be
// present in non-interactive mode.
func (ev *Evaler) aliasKeys() []string {
	editNs := ev.lookupSubNs("edit")
	if editNs == nil {
		return nil
	}
	abbrVar := editNs.IndexString("abbr")
	if abbrVar == nil {
		return nil
	}
	abbr, ok := abbrVar.Get().(vals.Map)
	if !ok {
		return nil
	}
	var keys []string
	for it := abbr.Iterator(); it.HasElem(); it.Next() {
		k, _ := it.Elem()
		if s, ok := k.(string); ok {
			keys = append(keys, s)
		}
	}
	return keys
}

// builtinNames returns special forms and builtin function names. Function
// names are stored with a ~ suffix in the namespace; it is stripped here so
// all names are bare.
func (ev *Evaler) builtinNames() []string {
	seen := map[string]bool{}
	var names []string
	add := func(name string) {
		trimmed := strings.TrimSuffix(name, FnSuffix)
		if !seen[trimmed] {
			seen[trimmed] = true
			names = append(names, trimmed)
		}
	}
	for name := range IsBuiltinSpecial {
		add(name)
	}
	ev.Builtin().IterateKeysString(add)
	return names
}

// functionNames returns user-defined function names (fn~ variables) from the
// global scope, without the ~ suffix.
func (ev *Evaler) functionNames() []string {
	var fns []string
	ev.Global().IterateKeysString(func(name string) {
		if strings.HasSuffix(name, FnSuffix) {
			fns = append(fns, strings.TrimSuffix(name, FnSuffix))
		}
	})
	return fns
}

// variableNames returns all variable names from the global scope and the
// builtin namespace, with suffixes (~ and :) stripped.
func (ev *Evaler) variableNames() []string {
	seen := map[string]bool{}
	var names []string
	add := func(name string) {
		trimmed := strings.TrimSuffix(name, FnSuffix)
		trimmed = strings.TrimSuffix(trimmed, NsSuffix)
		if !seen[trimmed] {
			seen[trimmed] = true
			names = append(names, trimmed)
		}
	}
	ev.Global().IterateKeysString(add)
	ev.Builtin().IterateKeysString(add)
	return names
}

// lookupSubNs finds a sub-namespace in the builtin namespace by name.
func (ev *Evaler) lookupSubNs(name string) *Ns {
	v := ev.Builtin().IndexString(name + NsSuffix)
	if v == nil {
		return nil
	}
	ns, ok := v.Get().(*Ns)
	if !ok {
		return nil
	}
	return ns
}
