CREATE TABLE IF NOT EXISTS actors (
    actor_id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    sex VARCHAR(10),
    date_of_birth DATE
);

ALTER TABLE actors OWNER TO postgres;

CREATE TABLE IF NOT EXISTS movies (
    movie_id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    release_date DATE,
    rating FLOAT
);

ALTER TABLE movies OWNER TO postgres;

CREATE TABLE IF NOT EXISTS movies_actors (
    movie_id INT REFERENCES movies(movie_id),
    actor_id INT REFERENCES actors(actor_id),
    PRIMARY KEY (movie_id, actor_id)
);

ALTER TABLE movies_actors OWNER TO postgres;

CREATE INDEX movie_name_index ON movies(name);
CREATE INDEX actor_name_index ON actors(name);

CREATE TABLE IF NOT EXISTS users (
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(50) NOT NULL,
    role VARCHAR(30) NOT NULL
);

ALTER TABLE users OWNER TO postgres;