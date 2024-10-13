package db

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Page struct {
	ID        int64           `json:"id"`
	Title     string          `json:"title"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Content   string          `json:"content"`
	Data      json.RawMessage `json:"data"`
	Deleted   bool            `json:"deleted"`
}

type Revision struct {
	ID              int64     `json:"id"`
	PageID          int64     `json:"page_id"`
	RevisionAt      time.Time `json:"revision_at"`
	PreviousContent string    `json:"previous_content"`
}

type PageManager struct {
	db                 *sql.DB
	createPageStmt     *sql.Stmt
	updatePageStmt     *sql.Stmt
	getPageByIDStmt    *sql.Stmt
	getPageByTitleStmt *sql.Stmt
	deletePageStmt     *sql.Stmt
	insertRevisionStmt *sql.Stmt
}

func NewPageManager(dbPath string) (*PageManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	pm := &PageManager{db: db}
	err = pm.initDB()
	if err != nil {
		return nil, err
	}
	pm.createPreparedStatements()
	return pm, err
}

func (pm *PageManager) initDB() error {
	createPagesTable := `
    CREATE TABLE IF NOT EXISTS pages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        content TEXT,
        data BLOB,
        deleted BOOLEAN DEFAULT 0
    );`
	createRevisionsTable := `
    CREATE TABLE IF NOT EXISTS revisions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        page_id INTEGER,
        revision_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        previous_content TEXT,
        FOREIGN KEY(page_id) REFERENCES pages(id)
    );`
	_, err := pm.db.Exec(createPagesTable)
	if err != nil {
		return err
	}
	_, err = pm.db.Exec(createRevisionsTable)
	return err
}

func (pm *PageManager) createPreparedStatements() error {
	createPageStmt, err := pm.db.Prepare(`
        INSERT INTO pages (title, content, data, created_at, updated_at)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	updatePageStmt, err := pm.db.Prepare(`
        UPDATE pages 
        SET title = ?, content = ?, data = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND deleted = 0`)
	if err != nil {
		return err
	}
	deletePageStmt, err := pm.db.Prepare(`
        UPDATE pages SET deleted = 1 WHERE id = ?`)
	if err != nil {
		return err
	}
	insertRevisionStmt, err := pm.db.Prepare(`
        INSERT INTO revisions (page_id, previous_content, revision_at)
        VALUES (?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	getPageByIdStmt, err := pm.db.Prepare(`
        SELECT id, title, created_at, updated_at, content, data, deleted 
        FROM pages WHERE id = ? AND deleted = 0`)
	if err != nil {
		return err
	}
	getPageByTitleStmt, err := pm.db.Prepare(`
        SELECT id, title, created_at, updated_at, content, data, deleted 
        FROM pages WHERE title = ? AND deleted = 0`)
	if err != nil {
		return err
	}
	pm.createPageStmt = createPageStmt
	pm.updatePageStmt = updatePageStmt
	pm.deletePageStmt = deletePageStmt
	pm.insertRevisionStmt = insertRevisionStmt
	pm.getPageByIDStmt = getPageByIdStmt
	pm.getPageByTitleStmt = getPageByTitleStmt
	return nil
}

func (pm *PageManager) CreatePage(title, content string, data json.RawMessage) (int64, error) {
	res, err := pm.createPageStmt.Exec(title, content, data)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (pm *PageManager) UpdatePage(id int64, newTitle, newContent string, newData json.RawMessage) error {
	// Save current page content to revisions
	page, err := pm.GetPageByID(id)
	if err != nil {
		return err
	}
	if err := pm.saveRevision(id, page.Content); err != nil {
		return err
	}

	_, err = pm.updatePageStmt.Exec(newTitle, newContent, newData, id)
	return err
}

func (pm *PageManager) GetPage(id string) (*Page, error) {
	val, err := strconv.ParseInt(id, 10, 64)
	if err == nil {
		return pm.GetPageByID(val)
	} else {
		return pm.GetPageByTitle(id)
	}
}

func (pm *PageManager) GetPageByID(id int64) (*Page, error) {
	row := pm.getPageByIDStmt.QueryRow(id)

	var page Page
	var rawData []byte
	err := row.Scan(&page.ID, &page.Title, &page.CreatedAt, &page.UpdatedAt, &page.Content, &rawData, &page.Deleted)
	if err != nil {
		return nil, err
	}
	page.Data = rawData
	return &page, nil
}

func (pm *PageManager) GetPageByTitle(title string) (*Page, error) {
	row := pm.getPageByTitleStmt.QueryRow(title)

	var page Page
	var rawData []byte
	err := row.Scan(&page.ID, &page.Title, &page.CreatedAt, &page.UpdatedAt, &page.Content, &rawData, &page.Deleted)
	if err != nil {
		return nil, err
	}
	page.Data = rawData
	return &page, nil
}

func (pm *PageManager) DeletePage(id int64) error {
	_, err := pm.deletePageStmt.Exec(id)
	return err
}

func (pm *PageManager) saveRevision(pageID int64, previousContent string) error {
	_, err := pm.insertRevisionStmt.Exec(pageID, previousContent)
	return err
}

func (pm *PageManager) Close() error {
	return pm.db.Close()
}
