CREATE TABLE IF NOT EXISTS predefined_avatars (
    id BIGSERIAL PRIMARY KEY,
    gender VARCHAR(16) NOT NULL CHECK (gender IN ('male','female','neutral')),
    url TEXT NOT NULL UNIQUE
);

INSERT INTO predefined_avatars (gender, url) VALUES
    ('female', 'https://skyfoxcinema.sirv.com/Avatars/female-1.jpg'),
    ('female', 'https://skyfoxcinema.sirv.com/Avatars/female-2.jpg'),
    ('female', 'https://skyfoxcinema.sirv.com/Avatars/female-3.jpg'),
    ('female', 'https://skyfoxcinema.sirv.com/Avatars/female-4.jpg'),
    ('female', 'https://skyfoxcinema.sirv.com/Avatars/female-5.jpg'),
    ('male',   'https://skyfoxcinema.sirv.com/Avatars/male-1.jpg'),
    ('male',   'https://skyfoxcinema.sirv.com/Avatars/male-2.jpg'),
    ('male',   'https://skyfoxcinema.sirv.com/Avatars/male-3.jpg'),
    ('male',   'https://skyfoxcinema.sirv.com/Avatars/male-4.jpg'),
    ('male',   'https://skyfoxcinema.sirv.com/Avatars/male-5.jpg'),
    ('neutral','https://skyfoxcinema.sirv.com/Avatars/neutral-1.jpg')
ON CONFLICT (url) DO NOTHING;
