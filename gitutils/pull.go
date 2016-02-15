package gitutils

import (
	"fmt"
	
	"gopkg.in/libgit2/git2go.v22"
)

type PullInfo struct {
	Reponame string
	Repo *git.Repository
	Branch string
	Signature *git.Signature
}

// Simulate Git pull
func (p *PullInfo ) GitPull() ( error ) {

	fmt.Print("Pulling\n")

	// fetch
	remotes,err := p.Repo.Remotes.List()
	if err != nil {
		return err
	} else {
		origin, err := p.Repo.Remotes.Lookup(remotes[0])
		defer origin.Free()
		remoteCallbacks := git.RemoteCallbacks{
			CredentialsCallback:  credentialsCallback,
			CertificateCheckCallback: certificateCheckCallback,
		}
		fetchOptions := &git.FetchOptions {
			RemoteCallbacks: remoteCallbacks,
		}
		// origin.SetCallbacks(remoteCallbacks)
		if err != nil {
			return err
		} else {
			refspec := make([]string, 0)
			fmt.Print("Fetching\n")
			err = origin.Fetch(refspec, fetchOptions, "pull")
			if err != nil {
				return err
			} 
		}
	}

	// merge
	local := fmt.Sprintf("refs/heads/%s", p.Branch)
	remote := fmt.Sprintf("refs/remotes/origin/%s", p.Branch)
	localref, err := p.Repo.References.Lookup(local)
	if err != nil {
		return err
	}
	remoteref, err := p.Repo.References.Lookup(remote)
	if err != nil {
		return err
	}
	mergeHead, err := p.Repo.AnnotatedCommitFromRef(remoteref)
	if err != nil {
		return err
	}
	mergeHeads := make([]*git.AnnotatedCommit,1)
	mergeHeads[0] = mergeHead
	mergeAnalysis, _ ,err := p.Repo.MergeAnalysis(mergeHeads)
	if err != nil {
		return err
	}
	if ( mergeAnalysis & git.MergeAnalysisUnborn) == git.MergeAnalysisUnborn {
		err = fmt.Errorf("Cannot merge an unborn commit.")
		return err
	}
	if ( mergeAnalysis & git.MergeAnalysisUpToDate) == git.MergeAnalysisUpToDate {
		fmt.Printf("Repo up to date\n")
		p.Repo.StateCleanup()
		return nil
	}
	fmt.Printf("Merging\n")
	err = p.Repo.Merge(mergeHeads,nil,&git.CheckoutOpts{Strategy: git.CheckoutUseTheirs})
	if err != nil {
		return err
	}

	// commit
	fmt.Printf("Committing\n")
	headIndex, _ := p.Repo.Index()
	headWriteOid, _ := headIndex.WriteTree()
	headTree, _ := p.Repo.LookupTree(headWriteOid)
	currentTip, _ := p.Repo.LookupCommit(localref.Target())
	branchTip, _ := p.Repo.LookupCommit(remoteref.Target())
	_,err = p.Repo.CreateCommit("HEAD", p.Signature, p.Signature, "merged "+p.Branch, headTree, currentTip, branchTip) 
	if err != nil {
		return err
	}

	// cleanup
	p.Repo.StateCleanup()
	return nil
}