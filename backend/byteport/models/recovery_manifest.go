package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type RecoveryManifest struct {
	Name string `yaml:"NAME"`
	Description string `yaml:"DESCRIPTION,omitempty"`
	Services []RecoveryManifestService `yaml:"SERVICES"`
}

type RecoveryManifestService struct {
	Name string `yaml:"NAME"`
	Path string `yaml:"PATH"`
	Port int `yaml:"PORT"`
	Runtime string `yaml:"RUNTIME,omitempty"`
	Build []string `yaml:"BUILD,omitempty"`
	Env map[string]string `yaml:"ENV,omitempty"`
}

func ParseRecoveryManifest(raw []byte, source SourceSnapshotID, path string) (ManifestRevision, RecoveryManifest, error) {
	sum:=sha256.Sum256(raw)
	rev:=ManifestRevision{
		ID: ManifestRevisionID(hex.EncodeToString(sum[:])),
		SourceSnapshot: source,
		Path:path,
		SchemaVersion:"recovery-v1",
		ContentDigest:hex.EncodeToString(sum[:]),
	}
	var m RecoveryManifest
	if err:=yaml.UnmarshalWithOptions(raw,&m,yaml.Strict());err!=nil{return rev,m,fmt.Errorf("parse manifest: %w",err)}
	if strings.TrimSpace(m.Name)=="" { return rev,m,fmt.Errorf("NAME is required") }
	if len(m.Services)==0 { return rev,m,fmt.Errorf("SERVICES must contain at least one service") }
	seen:=map[string]struct{}{}
	for i,s:=range m.Services {
		name:=strings.TrimSpace(s.Name)
		if name=="" { return rev,m,fmt.Errorf("SERVICES[%d].NAME is required",i) }
		if _,ok:=seen[name];ok{return rev,m,fmt.Errorf("duplicate service NAME %q",name)}
		seen[name]=struct{}{}
		if strings.TrimSpace(s.Path)=="" { return rev,m,fmt.Errorf("SERVICES[%d].PATH is required",i) }
		clean:=filepath.Clean(s.Path)
		if filepath.IsAbs(clean)||clean==".."||strings.HasPrefix(clean,".."+string(filepath.Separator)){return rev,m,fmt.Errorf("SERVICES[%d].PATH escapes repository",i)}
		if s.Port<1||s.Port>65535{return rev,m,fmt.Errorf("SERVICES[%d].PORT out of range",i)}
	}
	return rev,m,nil
}
