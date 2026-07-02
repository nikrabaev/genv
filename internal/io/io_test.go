package io_test

import (
	"os"
	"path/filepath"
	"testing"

	genvio "github.com/nikrabaev/genv/internal/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- WriteFileAtomic ---

func TestWriteFileAtomic_Basic(t *testing.T) {
	dir := t.TempDir()
	err := genvio.WriteFileAtomic(dir, "test.txt", []byte("hello"))
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(dir, "test.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestWriteFileAtomic_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	err := genvio.WriteFileAtomic(dir, "apps/api/.env", []byte("KEY=value"))
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(dir, "apps/api/.env"))
	require.NoError(t, err)
	assert.Equal(t, "KEY=value", string(data))
}

func TestWriteFileAtomic_NoTmpFileLeft(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, genvio.WriteFileAtomic(dir, "test.txt", []byte("hello")))
	_, err := os.Stat(filepath.Join(dir, "test.txt.genv-tmp"))
	assert.True(t, os.IsNotExist(err), "tmp file should be cleaned up after successful write")
}

func TestWriteFileAtomic_Overwrite(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, genvio.WriteFileAtomic(dir, "f.txt", []byte("v1")))
	require.NoError(t, genvio.WriteFileAtomic(dir, "f.txt", []byte("v2")))
	data, _ := os.ReadFile(filepath.Join(dir, "f.txt"))
	assert.Equal(t, "v2", string(data))
}

// --- FindRoot ---

func TestFindRoot_Direct(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "genv.json"), []byte("{}"), 0o644))
	got, ok := genvio.FindRoot(dir)
	assert.True(t, ok)
	assert.Equal(t, dir, got)
}

func TestFindRoot_Nested(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "genv.json"), []byte("{}"), 0o644))
	sub := filepath.Join(dir, "apps/api")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	got, ok := genvio.FindRoot(sub)
	assert.True(t, ok)
	assert.Equal(t, dir, got)
}

func TestFindRoot_NotFound(t *testing.T) {
	dir := t.TempDir() // no genv.json anywhere above
	_, ok := genvio.FindRoot(dir)
	assert.False(t, ok)
}
