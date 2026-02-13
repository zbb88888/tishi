package datastore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// Store provides read/write access to the data/ directory.
type Store struct {
	dataDir string
	log     *zap.Logger
}

// NewStore creates a Store rooted at dataDir.
func NewStore(dataDir string, log *zap.Logger) *Store {
	return &Store{
		dataDir: dataDir,
		log:     log,
	}
}

// DataDir returns the root data directory path.
func (s *Store) DataDir() string {
	return s.dataDir
}

// --- Projects ---

// projectsDir returns the path to data/projects/.
func (s *Store) projectsDir() string {
	return filepath.Join(s.dataDir, "projects")
}

// ProjectFilename converts full_name (owner/repo) to filename (owner__repo.json).
func ProjectFilename(fullName string) string {
	return strings.ReplaceAll(fullName, "/", "__") + ".json"
}

// ProjectIDFromFullName converts owner/repo to owner__repo.
func ProjectIDFromFullName(fullName string) string {
	return strings.ReplaceAll(fullName, "/", "__")
}

// LoadProject reads a single project JSON file.
func (s *Store) LoadProject(id string) (*Project, error) {
	path := filepath.Join(s.projectsDir(), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading project %s: %w", id, err)
	}

	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing project %s: %w", id, err)
	}
	return &p, nil
}

// SaveProject writes a project JSON file atomically.
func (s *Store) SaveProject(p *Project) error {
	dir := s.projectsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating projects dir: %w", err)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling project %s: %w", p.ID, err)
	}
	data = append(data, '\n')

	path := filepath.Join(dir, p.ID+".json")
	// Atomic write via temp file + rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// ListProjects reads all project JSON files from data/projects/.
func (s *Store) ListProjects() ([]*Project, error) {
	dir := s.projectsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing projects dir: %w", err)
	}

	var projects []*Project
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		p, err := s.LoadProject(id)
		if err != nil {
			s.log.Warn("跳过无效项目文件", zap.String("file", e.Name()), zap.Error(err))
			continue
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// --- Snapshots ---

func (s *Store) snapshotsDir() string {
	return filepath.Join(s.dataDir, "snapshots")
}

// AppendSnapshot appends a snapshot line to data/snapshots/{date}.jsonl.
func (s *Store) AppendSnapshot(snap *Snapshot) error {
	dir := s.snapshotsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating snapshots dir: %w", err)
	}

	line, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("marshaling snapshot: %w", err)
	}
	line = append(line, '\n')

	path := filepath.Join(dir, snap.Date+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening snapshot file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("writing snapshot line: %w", err)
	}

	return nil
}

// LoadSnapshots reads all snapshot lines for a given date.
func (s *Store) LoadSnapshots(date string) ([]*Snapshot, error) {
	path := filepath.Join(s.snapshotsDir(), date+".jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening snapshot file: %w", err)
	}
	defer f.Close()

	var snaps []*Snapshot
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var snap Snapshot
		if err := json.Unmarshal(line, &snap); err != nil {
			s.log.Warn("跳过无效快照行", zap.String("date", date), zap.Error(err))
			continue
		}
		snaps = append(snaps, &snap)
	}

	return snaps, scanner.Err()
}

// --- Rankings ---

func (s *Store) rankingsDir() string {
	return filepath.Join(s.dataDir, "rankings")
}

// SaveRanking writes a ranking JSON file.
func (s *Store) SaveRanking(r *Ranking) error {
	dir := s.rankingsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating rankings dir: %w", err)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling ranking: %w", err)
	}
	data = append(data, '\n')

	path := filepath.Join(dir, r.Date+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// LoadRanking reads a ranking JSON file for a given date.
func (s *Store) LoadRanking(date string) (*Ranking, error) {
	path := filepath.Join(s.rankingsDir(), date+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading ranking %s: %w", date, err)
	}

	var r Ranking
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parsing ranking %s: %w", date, err)
	}
	return &r, nil
}

// LoadLatestRanking finds and reads the most recent ranking file.
func (s *Store) LoadLatestRanking() (*Ranking, error) {
	dir := s.rankingsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing rankings dir: %w", err)
	}

	// JSON files sorted by name (date) descending
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	if len(files) == 0 {
		return nil, nil
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	return s.LoadRanking(files[0])
}

// --- Posts ---

func (s *Store) postsDir() string {
	return filepath.Join(s.dataDir, "posts")
}

// SavePost writes a post JSON file.
func (s *Store) SavePost(p *Post) error {
	dir := s.postsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating posts dir: %w", err)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling post %s: %w", p.Slug, err)
	}
	data = append(data, '\n')

	path := filepath.Join(dir, p.Slug+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// ListPosts reads all post JSON files.
func (s *Store) ListPosts() ([]*Post, error) {
	dir := s.postsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing posts dir: %w", err)
	}

	var posts []*Post
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			s.log.Warn("读取文章失败", zap.String("file", e.Name()), zap.Error(err))
			continue
		}
		var p Post
		if err := json.Unmarshal(data, &p); err != nil {
			s.log.Warn("解析文章失败", zap.String("slug", slug), zap.Error(err))
			continue
		}
		posts = append(posts, &p)
	}

	return posts, nil
}

// --- Categories ---

// LoadCategories reads data/categories.json.
func (s *Store) LoadCategories() ([]Category, error) {
	path := filepath.Join(s.dataDir, "categories.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading categories.json: %w", err)
	}

	var cats []Category
	if err := json.Unmarshal(data, &cats); err != nil {
		return nil, fmt.Errorf("parsing categories.json: %w", err)
	}
	return cats, nil
}
