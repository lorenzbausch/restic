package khepri_test

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fd0/khepri"
	"github.com/fd0/khepri/backend"
)

var testPassword = "foobar"
var testCleanup = flag.Bool("test.cleanup", true, "clean up after running tests (remove local backend directory with all content)")

func setupBackend(t testing.TB) *backend.Local {
	tempdir, err := ioutil.TempDir("", "khepri-test-")
	ok(t, err)

	b, err := backend.CreateLocal(tempdir)
	ok(t, err)

	return b
}

func teardownBackend(t testing.TB, b *backend.Local) {
	if !*testCleanup {
		t.Logf("leaving local backend at %s\n", b.Location())
		return
	}

	ok(t, os.RemoveAll(b.Location()))
}

func setupKey(t testing.TB, be backend.Server, password string) *khepri.Key {
	k, err := khepri.CreateKey(be, password)
	ok(t, err)

	return k
}

func TestRepo(t *testing.T) {
	be := setupBackend(t)
	defer teardownBackend(t, be)
	_ = setupKey(t, be, testPassword)
}

func TestEncryptDecrypt(t *testing.T) {
	be := setupBackend(t)
	defer teardownBackend(t, be)
	k := setupKey(t, be, testPassword)

	for _, size := range []int{5, 23, 1 << 20, 7<<20 + 123} {
		data := make([]byte, size)
		f, err := os.Open("/dev/urandom")
		ok(t, err)

		_, err = io.ReadFull(f, data)
		ok(t, err)

		ciphertext := khepri.GetChunkBuf("TestEncryptDecrypt")
		n, err := k.Encrypt(ciphertext, data)
		ok(t, err)

		plaintext, err := k.Decrypt(ciphertext[:n])
		ok(t, err)

		khepri.FreeChunkBuf("TestEncryptDecrypt", ciphertext)

		equals(t, plaintext, data)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	size := 8 << 20 // 8MiB
	data := make([]byte, size)

	be := setupBackend(b)
	defer teardownBackend(b, be)
	k := setupKey(b, be, testPassword)

	b.ResetTimer()
	b.SetBytes(int64(size))

	buf := khepri.GetChunkBuf("BenchmarkEncrypt")
	for i := 0; i < b.N; i++ {
		_, err := k.Encrypt(buf, data)
		ok(b, err)
	}
	khepri.FreeChunkBuf("BenchmarkEncrypt", buf)
}

func BenchmarkDecrypt(b *testing.B) {
	size := 8 << 20 // 8MiB
	data := make([]byte, size)

	be := setupBackend(b)
	defer teardownBackend(b, be)
	k := setupKey(b, be, testPassword)

	ciphertext := khepri.GetChunkBuf("BenchmarkDecrypt")
	n, err := k.Encrypt(ciphertext, data)
	ok(b, err)

	b.ResetTimer()
	b.SetBytes(int64(size))

	for i := 0; i < b.N; i++ {
		_, err := k.Decrypt(ciphertext[:n])
		ok(b, err)
	}
	khepri.FreeChunkBuf("BenchmarkDecrypt", ciphertext)
}