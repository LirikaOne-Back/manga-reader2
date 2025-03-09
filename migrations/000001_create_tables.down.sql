-- migrations/000001_create_tables.down.sql

-- Удаление индексов
DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_comments_chapter_id;
DROP INDEX IF EXISTS idx_comments_manga_id;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_reading_history_chapter_id;
DROP INDEX IF EXISTS idx_reading_history_manga_id;
DROP INDEX IF EXISTS idx_reading_history_user_id;
DROP INDEX IF EXISTS idx_bookmarks_manga_id;
DROP INDEX IF EXISTS idx_bookmarks_user_id;
DROP INDEX IF EXISTS idx_chapter_views_user_id;
DROP INDEX IF EXISTS idx_chapter_views_chapter_id;
DROP INDEX IF EXISTS idx_manga_views_user_id;
DROP INDEX IF EXISTS idx_manga_views_manga_id;
DROP INDEX IF EXISTS idx_pages_number;
DROP INDEX IF EXISTS idx_pages_chapter_id;
DROP INDEX IF EXISTS idx_chapters_number;
DROP INDEX IF EXISTS idx_chapters_manga_id;
DROP INDEX IF EXISTS idx_manga_status;
DROP INDEX IF EXISTS idx_manga_title;

-- Удаление таблиц в порядке зависимостей
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS reading_history;
DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS chapter_views;
DROP TABLE IF EXISTS manga_views;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS chapters;
DROP TABLE IF EXISTS manga_genres;
DROP TABLE IF EXISTS manga;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS users;