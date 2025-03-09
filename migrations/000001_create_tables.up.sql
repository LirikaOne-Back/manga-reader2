-- migrations/000001_create_tables.up.sql

-- Создание расширения для UUID (если понадобится)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица с жанрами
CREATE TABLE IF NOT EXISTS genres (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица с мангой
CREATE TABLE IF NOT EXISTS manga (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_image VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'ongoing', -- ongoing, completed, hiatus
    author VARCHAR(100),
    artist VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица связи манги и жанров
CREATE TABLE IF NOT EXISTS manga_genres (
    manga_id INTEGER NOT NULL,
    genre_id INTEGER NOT NULL,
    PRIMARY KEY (manga_id, genre_id),
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES genres(id) ON DELETE CASCADE
);

-- Таблица с главами
CREATE TABLE IF NOT EXISTS chapters (
    id SERIAL PRIMARY KEY,
    manga_id INTEGER NOT NULL,
    number NUMERIC(8,2) NOT NULL, -- Поддержка глав типа 1.5, 2.3 и т.д.
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
    UNIQUE (manga_id, number)
);

-- Таблица со страницами
CREATE TABLE IF NOT EXISTS pages (
    id SERIAL PRIMARY KEY,
    chapter_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    image_path VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    UNIQUE (chapter_id, number)
);

-- Таблица с пользователями
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- user, admin
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица просмотров манги
CREATE TABLE IF NOT EXISTS manga_views (
    id SERIAL PRIMARY KEY,
    manga_id INTEGER NOT NULL,
    user_id INTEGER,
    ip_address VARCHAR(50),
    viewed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Таблица просмотров глав
CREATE TABLE IF NOT EXISTS chapter_views (
    id SERIAL PRIMARY KEY,
    chapter_id INTEGER NOT NULL,
    user_id INTEGER,
    ip_address VARCHAR(50),
    viewed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Таблица закладок
CREATE TABLE IF NOT EXISTS bookmarks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    manga_id INTEGER NOT NULL,
    chapter_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    UNIQUE (user_id, manga_id)
);

-- Таблица истории чтения
CREATE TABLE IF NOT EXISTS reading_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    manga_id INTEGER NOT NULL,
    chapter_id INTEGER NOT NULL,
    page_number INTEGER,
    read_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
);

-- Таблица комментариев
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    manga_id INTEGER,
    chapter_id INTEGER,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE,
    CHECK ((manga_id IS NOT NULL AND chapter_id IS NULL) OR (manga_id IS NULL AND chapter_id IS NOT NULL))
);

-- Таблица сессий
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    refresh_token VARCHAR(255) NOT NULL,
    user_agent VARCHAR(255),
    ip_address VARCHAR(50),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_manga_title ON manga(title);
CREATE INDEX idx_manga_status ON manga(status);
CREATE INDEX idx_chapters_manga_id ON chapters(manga_id);
CREATE INDEX idx_chapters_number ON chapters(number);
CREATE INDEX idx_pages_chapter_id ON pages(chapter_id);
CREATE INDEX idx_pages_number ON pages(number);
CREATE INDEX idx_manga_views_manga_id ON manga_views(manga_id);
CREATE INDEX idx_manga_views_user_id ON manga_views(user_id);
CREATE INDEX idx_chapter_views_chapter_id ON chapter_views(chapter_id);
CREATE INDEX idx_chapter_views_user_id ON chapter_views(user_id);
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_manga_id ON bookmarks(manga_id);
CREATE INDEX idx_reading_history_user_id ON reading_history(user_id);
CREATE INDEX idx_reading_history_manga_id ON reading_history(manga_id);
CREATE INDEX idx_reading_history_chapter_id ON reading_history(chapter_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_manga_id ON comments(manga_id);
CREATE INDEX idx_comments_chapter_id ON comments(chapter_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);