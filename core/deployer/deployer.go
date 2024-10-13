package deployer

import "lamprey/core/db"

type Deployer interface {
	DeployArticle(page db.Page) error
	DeployData(page db.Page) error
}

func GetPageHtml(page db.Page) (string, error) {
	return "", nil
}
