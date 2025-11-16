package converter

import (
	serviceModel "PR/internal/model"
	repoModel "PR/internal/repository/pr/model"
)

func FromRepo(pr *repoModel.PullRequest) *serviceModel.PullRequest {
	return &serviceModel.PullRequest{
		ID:                pr.ID,
		Name:              pr.Name,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
		AuthorID:          pr.AuthorID,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func FromRepoShort(pr *repoModel.PullRequestShort) *serviceModel.PullRequestShort {
	return &serviceModel.PullRequestShort{
		ID:       pr.ID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   pr.Status,
	}
}

func FromRepoShortList(prs []*repoModel.PullRequestShort) []*serviceModel.PullRequestShort {
	prList := make([]*serviceModel.PullRequestShort, 0, len(prs))
	for _, p := range prs {
		prList = append(prList, FromRepoShort(p))
	}
	return prList
}
