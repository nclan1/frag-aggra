
CREATE TABLE posts (
    -- The primary key for this table, an auto-incrementing integer.
    id SERIAL PRIMARY KEY,
    
    -- The unique ID from Reddit itself (e.g., 't3_xyz123'), prevents duplicates.
    reddit_id VARCHAR(20) UNIQUE NOT NULL,
    
    -- The full URL to the post.
    url TEXT UNIQUE NOT NULL,
    
    -- The username of the seller.
    seller_username VARCHAR(255),
    
    -- A timestamp automatically set to when the row was created.
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE listings (
    -- The primary key for this table.
    id SERIAL PRIMARY KEY,
    
    -- The Foreign Key linking to the 'posts' table.
    -- If a post is deleted, all its associated listings will also be deleted.
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    
    -- The name of the perfume being sold.
    name VARCHAR(255) NOT NULL,
    
    -- The size of the bottle (e.g., '100ml', '50/100ml').
    size VARCHAR(50),
    
    -- The price of the perfume (e.g., '$120').
    price VARCHAR(50),
    
    -- A timestamp automatically set to when the row was created.
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- In the up migration, after creating tables:
CREATE INDEX idx_listings_post_id ON listings(post_id);
CREATE INDEX idx_posts_reddit_id ON posts(reddit_id);
CREATE INDEX idx_posts_seller_username ON posts(seller_username);