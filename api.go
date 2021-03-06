package easygit

import (
	"strings"
	"time"

	"github.com/libgit2/git2go"
)

// Add / Commit
// --------------------------------------------------------

func AddAll(repoPath string) error {

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	idx, err := repo.Index()
	if err != nil {
		return err
	}

	err = idx.AddAll([]string{}, git.IndexAddDefault, nil)
	if err != nil {
		return err
	}

	err = idx.Write()

	return err
}

func Commit(repoPath string, message string, name string, email string) error {

	sig := &git.Signature{Name: name, Email: email, When: time.Now()}

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	idx, err := repo.Index()
	if err != nil {
		return err
	}

	treeID, err := idx.WriteTree()
	if err != nil {
		return err
	}

	tree, err := repo.LookupTree(treeID)
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if head == nil {

		_, err = repo.CreateCommit("HEAD", sig, sig, message, tree)
		return err

	} else if err != nil {

		return err

	} else {

		parent, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		_, err = repo.CreateCommit("HEAD", sig, sig, message, tree, parent)
		return err

	}

}

// Branches
// --------------------------------------------------------

func ListBranches(repoPath string) ([]string, error) {

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	iter, err := repo.NewBranchIterator(git.BranchLocal)
	if err != nil {
		return nil, err
	}

	var branches []string

	branch, _, err := iter.Next()
	for err == nil {
		name, _ := branch.Name()
		branches = append(branches, name)
		branch, _, err = iter.Next()
	}

	return branches, nil
}

func CheckoutBranch(repoPath string, branchName string) error {

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return err
	}

	commit, err := repo.LookupCommit(branch.Target())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	err = repo.CheckoutTree(tree, &git.CheckoutOpts{})
	if err != nil {
		return err
	}

	err = repo.SetHead("refs/heads/" + branchName)
	if err != nil {
		return err
	}

	return nil
}

func PushBranch(repoPath string, remoteName string, branch string, user string, pass string) error {

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return nil
	}

	remote, err := repo.Remotes.Lookup(remoteName)
	if err != nil {
		return nil
	}

	err = remote.Push([]string{"refs/heads/" + branch}, &git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback: func(url string, username_from_url string, allowed_types git.CredType) (git.ErrorCode, *git.Cred) {
				_, creds := git.NewCredUserpassPlaintext(user, pass)
				return git.ErrOk, &creds
			},
		},
	})

	return err
}

func CreateBranch(repoPath string, from string, to string) error {

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	fromBranch, err := repo.LookupBranch(from, git.BranchLocal)
	if err != nil {
		return err
	}

	fromCommit, err := repo.LookupCommit(fromBranch.Target())
	if err != nil {
		return err
	}

	_, err = repo.CreateBranch(to, fromCommit, false)
	if err != nil {
		return err
	}

	return nil
}

func CurrentBranch(repoPath string) (string, error) {

	repo, repoErr := git.OpenRepository(repoPath)
	if repoErr != nil {
		return "", repoErr
	}

	head, headErr := repo.Head()
	if repoErr != nil {
		return "", headErr
	}

	return strings.Split(head.Name(), "/")[2], nil
}

func DeleteBranch(repoPath string, branchName string) error {

	repo, repoErr := git.OpenRepository(repoPath)
	if repoErr != nil {
		return repoErr
	}

	branch, branchErr := repo.LookupBranch(branchName, git.BranchLocal)
	if branchErr != nil {
		return branchErr
	}
	return branch.Delete()
}
