package easygit

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/libgit2/git2go"
)

func TestCommit(t *testing.T) {

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	err := AddAll(repo.Workdir())
	checkFatal(t, err)

	err = Commit(repo.Workdir(), "First commit", "First Last", "first@last.com")
	checkFatal(t, err)

	head, err := repo.Head()
	checkFatal(t, err)

	commit, err := repo.LookupCommit(head.Target())
	checkFatal(t, err)

	if commit.Message() != "First commit" {
		fail(t)
	}

	tree, err := commit.Tree()
	checkFatal(t, err)

	file := tree.EntryByName("README")
	if file == nil {
		fail(t)
	}

}

func TestCommitOtherBranch(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	defer cleanupTestRepo(t, repo)

	CreateBranch(repo.Workdir(), "master", "slave")
	err := CheckoutBranch(repo.Workdir(), "slave")
	checkFatal(t, err)

	err = ioutil.WriteFile(repo.Workdir()+"/SLAVEFILE", []byte("foo\n"), 0644)
	checkFatal(t, err)

	AddAll(repo.Workdir())
	Commit(repo.Workdir(), "Slave commit", "First Last", "first@last.com")

	head, err := repo.Head()
	checkFatal(t, err)
	commit, err := repo.LookupCommit(head.Target())
	checkFatal(t, err)

	if commit.Message() != "Slave commit" {
		fail(t)
	}

	tree, err := commit.Tree()
	checkFatal(t, err)

	file := tree.EntryByName("SLAVEFILE")
	if file == nil {
		fail(t)
	}

	file = tree.EntryByName("README")
	if file == nil {
		fail(t)
	}

	branch, err := CurrentBranch(repo.Workdir())
	checkFatal(t, err)
	if branch != "slave" {
		fail(t)
	}

}

func TestAddAll(t *testing.T) {

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	err := AddAll(repo.Workdir())
	checkFatal(t, err)

	index, err := repo.Index()
	checkFatal(t, err)

	entry, err := index.EntryByPath("README", 0)
	checkFatal(t, err)

	if entry == nil {
		fail(t)
	}
}

func TestListBranches(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	defer cleanupTestRepo(t, repo)

	branches, err := ListBranches(repo.Workdir())
	checkFatal(t, err)
	if branches[0] != "master" {
		fail(t)
	}
}

func TestCreateBranch(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	defer cleanupTestRepo(t, repo)

	CreateBranch(repo.Workdir(), "master", "slave")

	branches, err := ListBranches(repo.Workdir())
	checkFatal(t, err)
	if branches[0] != "master" || branches[1] != "slave" {
		fail(t)
	}

}

func TestCheckoutBranch(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	defer cleanupTestRepo(t, repo)

	CreateBranch(repo.Workdir(), "master", "slave")
	err := CheckoutBranch(repo.Workdir(), "slave")
	checkFatal(t, err)
	branch, err := CurrentBranch(repo.Workdir())
	checkFatal(t, err)
	if branch != "slave" {
		fail(t)
	}

}

func TestDeleteBranch(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	defer cleanupTestRepo(t, repo)

	CreateBranch(repo.Workdir(), "master", "slave")

	err := DeleteBranch(repo.Workdir(), "slave")
	checkFatal(t, err)

	branches, err := ListBranches(repo.Workdir())
	checkFatal(t, err)
	if len(branches) != 1 {
		fail(t)
	}

}

func TestPushBranch(t *testing.T) {
	t.Parallel()
	remoteRepo := createBareTestRepo(t)
	defer cleanupTestRepo(t, remoteRepo)
	localRepo := createTestRepo(t)
	defer cleanupTestRepo(t, localRepo)

	_, err := localRepo.Remotes.Create("test_push", remoteRepo.Path())
	checkFatal(t, err)

	seedTestRepo(t, localRepo)

	PushBranch(localRepo.Workdir(), "test_push", "master", "not", "used")

	_, err = localRepo.References.Lookup("refs/remotes/test_push/master")
	checkFatal(t, err)

	_, err = remoteRepo.References.Lookup("refs/heads/master")
	checkFatal(t, err)
}

func TestCurrentBranch(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	defer cleanupTestRepo(t, repo)

	branch, err := CurrentBranch(repo.Workdir())
	checkFatal(t, err)
	if branch != "master" {
		fail(t)
	}
}

// Test setup
// ---------------------------------------------------

func createTestRepo(t *testing.T) *git.Repository {
	path, err := ioutil.TempDir("", "easygit")
	checkFatal(t, err)
	repo, err := git.InitRepository(path, false)
	checkFatal(t, err)

	tmpfile := "README"
	err = ioutil.WriteFile(path+"/"+tmpfile, []byte("foo\n"), 0644)
	checkFatal(t, err)
	return repo
}

func createBareTestRepo(t *testing.T) *git.Repository {
	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)
	repo, err := git.InitRepository(path, true)
	checkFatal(t, err)
	return repo
}

func cleanupTestRepo(t *testing.T, r *git.Repository) {
	var err error
	if r.IsBare() {
		err = os.RemoveAll(r.Path())
	} else {
		err = os.RemoveAll(r.Workdir())
	}
	checkFatal(t, err)
	r.Free()
}

func seedTestRepo(t *testing.T, repo *git.Repository) (*git.Oid, *git.Oid) {
	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	err = idx.Write()
	checkFatal(t, err)
	treeID, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)
	commitID, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	return commitID, treeID
}

func fail(t *testing.T) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Fail at %v:%v; %v", file, line)
}

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}
