package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobSpec is a superset of the upstream spec, but
// we do not import it as importing test-infra is a
// massive hassle.
type JobSpec struct {
	Type    ProwJobType `json:"type,omitempty"`
	Job     string      `json:"job,omitempty"`
	BuildId string      `json:"buildid,omitempty"`

	Refs Refs `json:"refs,omitempty"`

	// rawSpec is the serialized form of the Spec
	rawSpec string

	// these fields allow the job to be targeted at a location
	identifer     string
	namespace     string
	baseNamespace string

	// if set, any new artifacts will be a child of this object
	owner *meta.OwnerReference
}

type ProwJobType string

const (
	PresubmitJob  ProwJobType = "presubmit"
	PostsubmitJob             = "postsubmit"
	PeriodicJob               = "periodic"
	BatchJob                  = "batch"
)

type Pull struct {
	Number int    `json:"number,omitempty"`
	Author string `json:"author,omitempty"`
	SHA    string `json:"sha,omitempty"`
}

type Refs struct {
	Org  string `json:"org,omitempty"`
	Repo string `json:"repo,omitempty"`

	BaseRef string `json:"base_ref,omitempty"`
	BaseSHA string `json:"base_sha,omitempty"`

	Pulls []Pull `json:"pulls,omitempty"`
}

func (r Refs) String() string {
	rs := []string{fmt.Sprintf("%s:%s", r.BaseRef, r.BaseSHA)}
	for _, pull := range r.Pulls {
		rs = append(rs, fmt.Sprintf("%d:%s", pull.Number, pull.SHA))
	}
	return strings.Join(rs, ",")
}

func (s *JobSpec) Namespace() string {
	return s.namespace
}

func (s *JobSpec) Owner() *meta.OwnerReference {
	return s.owner
}

func (s *JobSpec) SetNamespace(ns string) {
	s.namespace = ns
}

func (s *JobSpec) SetBaseNamespace(ns string) {
	s.baseNamespace = ns
}

func (s *JobSpec) SetOwner(owner *meta.OwnerReference) {
	s.owner = owner
}

// RefSpec returns a string that represents the canonical form of
// the references for use in uniqueness tests.
func (s *JobSpec) RefSpec() string {
	spec := &JobSpec{
		Refs: s.Refs,
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// ResolveSpecFromEnv will determine the Refs being
// tested in by parsing Prow environment variable contents
func ResolveSpecFromEnv() (*JobSpec, error) {
	specEnv, ok := os.LookupEnv("JOB_SPEC")
	if !ok {
		return nil, errors.New("$JOB_SPEC unset")
	}

	spec := &JobSpec{}
	if err := json.Unmarshal([]byte(specEnv), spec); err != nil {
		return nil, fmt.Errorf("malformed $JOB_SPEC: %v", err)
	}

	spec.rawSpec = specEnv

	return spec, nil
}
