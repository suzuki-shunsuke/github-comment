package config

import (
	"errors"

	"github.com/suzuki-shunsuke/github-comment/pkg/domain"
)

type Complement struct {
	PR   []domain.ComplementEntry
	Org  []domain.ComplementEntry
	Repo []domain.ComplementEntry
	SHA1 []domain.ComplementEntry
}

type rawComplement struct {
	PR   []map[string]interface{}
	Org  []map[string]interface{}
	Repo []map[string]interface{}
	SHA1 []map[string]interface{}
}

func convComplementEntries(maps []map[string]interface{}) ([]domain.ComplementEntry, error) {
	entries := make([]domain.ComplementEntry, len(maps))
	for i, m := range maps {
		entry, err := convComplementEntry(m)
		if err != nil {
			return nil, err
		}
		entries[i] = entry
	}
	return entries, nil
}

func convComplementEntry(m map[string]interface{}) (domain.ComplementEntry, error) {
	t, ok := m["type"]
	if !ok {
		return nil, errors.New(`"type" is required`)
	}
	typ, ok := t.(string)
	if !ok {
		return nil, errors.New(`"type" must be string`)
	}
	switch typ {
	case "envsubst":
		entry := ComplementEnvsubstEntry{}
		if err := newComplementEnvsubstEntry(m, &entry); err != nil {
			return nil, err
		}
		return &entry, nil
	case "template":
		entry := ComplementTemplateEntry{}
		if err := newComplementTemplateEntry(m, &entry); err != nil {
			return nil, err
		}
		return &entry, nil
	default:
		return nil, errors.New(`unsupported type: ` + typ)
	}
}

func (cpl *Complement) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val rawComplement
	if err := unmarshal(&val); err != nil {
		return err
	}

	pr, err := convComplementEntries(val.PR)
	if err != nil {
		return err
	}
	cpl.PR = pr

	org, err := convComplementEntries(val.Org)
	if err != nil {
		return err
	}
	cpl.Org = org

	repo, err := convComplementEntries(val.Repo)
	if err != nil {
		return err
	}
	cpl.Repo = repo

	sha, err := convComplementEntries(val.SHA1)
	if err != nil {
		return err
	}
	cpl.SHA1 = sha

	return nil
}
