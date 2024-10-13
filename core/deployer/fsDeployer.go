package deployer

import (
	"lamprey/core"
	"lamprey/core/db"
	"os"
	"path/filepath"
)

type FsDeployer struct {
	path string
}

func InitFsDeployer(config core.DeployToFolderConfig) FsDeployer {
	return FsDeployer{path: config.Path}
}

func (deployer FsDeployer) DeployArticle(page db.Page) error {
	content, err := GetPageHtml(page)
	if err != nil {
		return err
	}
	file, err := deployer.open(page, "html")
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func (deployer FsDeployer) DeployData(page db.Page) error {
	return nil
}

func (deployer FsDeployer) open(page db.Page, format string) (*os.File, error) {
	return os.OpenFile(
		filepath.Join(deployer.path, filepath.Clean(page.Title)+"."+format),
		os.O_RDWR|os.O_CREATE,
		0666)
}
