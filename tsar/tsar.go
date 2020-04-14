package tsar

import (
	"log"
	"errors"
	"fmt"

	"gopkg.in/libgit2/git2go.v22"
	"github.com/searchspring/repo-tsar/gitutils"
	"github.com/searchspring/repo-tsar/fileutils"
	"github.com/searchspring/repo-tsar/config"
)

type semaphore chan error

type RepoTsar struct{
	Config config.Config
	Branch string
	ReposList []string
	Signature *git.Signature
}

func (r *RepoTsar) Run() error {
	reposlist := r.ReposList 
	// if the reposlist is empty append all repos from config to reposlist
	c := r.Config.Repos
	if reposlist[0] == "" {
		// delete item from array
		reposlist = append(reposlist[:0], reposlist[0+1:]...)
		for k := range c {
			reposlist = append(reposlist,k)
		}
	}

	// Semaphore for concurrency
	thrnum := len(reposlist)
	sem := make(semaphore, thrnum)
	for _, k := range reposlist {
	//	go func(k string) error {

			_,ok := c[k]
			if ! ok {
				err := errors.New(fmt.Sprintf("Repo %#v is not defined in config\n", k))
				sem <- err
				return err
			}
			log.Printf("[%s, url: %s, path: %s, branch: %s]", k, c[k].URL, c[k].Path, c[k].Branch)
	
			// Createpath 
			path,err := fileutils.CreatePath(c[k].Path)
			if err != nil {
				sem <-err
				return fmt.Errorf("Cannot create path %#v : %s", c[k].Path,err)
			}
			
			// Clone Repo
			cloneinfo := &gitutils.CloneInfo{
				Reponame: k,
				Path: path,
				URL: c[k].URL,
				Branch: c[k].Branch,
			}
			repo, err := cloneinfo.CloneRepo()
			if err != nil {
				sem <-err
				return fmt.Errorf("Cannot clone repo %#v : %s", k,err)
			}
			
			// Git Pull
			pullinfo := gitutils.PullInfo{
				Reponame: k,
				Repo: repo,
				Branch: c[k].Branch,
				Signature: r.Signature,
			}
			err = pullinfo.GitPull()
			if err != nil {
				sem <-err
				return fmt.Errorf("Cannot pull repo %#v : %s ", k,err)
			}
	
			// If branch option, branch and checkout selected repos
			if r.Branch != "" {
				branchinfo := &gitutils.BranchInfo{
					Reponame: k,
					Branchname: r.Branch,
					Msg: "RepoTsar Branching",
					Repo: *repo,
					Signature: r.Signature,
				}
				err = branchinfo.GitBranch()
				if err != nil {
					sem <-err
					return fmt.Errorf("Cannot branch repo %#v : %s", k,err)
				}		
			}
			sem <-nil
			// return nil
	//	}(key)
	}

	// // Wait for threads to finish
	// for i := 0; i < thrnum; i++ { 
	// 	err := <-sem
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
