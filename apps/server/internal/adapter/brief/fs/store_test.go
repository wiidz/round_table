package fs

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleBrief = `meta:
  title: 测试模板
  description: 单元测试
topic: 示例主题
brief:
  goal: 形成共识
  agenda:
    - 议题 A
  in_scope: 范围
meeting:
  mode: decision
  max_rounds: 2
`

func TestStoreListReadWrite(t *testing.T) {
	root := t.TempDir()
	templates := t.TempDir()
	id := "test-template"
	if err := os.MkdirAll(filepath.Join(templates, id), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, id, "BRIEF.yaml"), []byte(sampleBrief), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStore(root, templates)
	list, err := s.ListTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != id || list[0].Source != "builtin" {
		t.Fatalf("list: %+v", list)
	}

	detail, err := s.ReadTemplate(id)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Launch.Topic != "示例主题" || detail.Launch.Brief.Goal != "形成共识" {
		t.Fatalf("launch: %+v", detail.Launch)
	}

	custom := `meta:
  title: 自定义
brief:
  goal: 自定义目标
`
	if err := s.WriteTemplate("my-brief", []byte(custom)); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteTemplate(id, []byte(custom)); err == nil {
		t.Fatal("expected builtin readonly error")
	}
}

func TestCloneFromMeetingDoc(t *testing.T) {
	s := NewStore(t.TempDir(), t.TempDir())
	doc := `# 会议简报 · Meeting Brief

## 会议主题

平衡调整

## 会议目标

输出方案

## 讨论议题

1. 机制 A
2. 机制 B

## 讨论范围

仅限 PVP
`
	draft, err := s.CloneFromMeetingDoc(doc)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Topic != "平衡调整" || len(draft.Brief.Agenda) != 2 {
		t.Fatalf("draft: %+v", draft)
	}
}
