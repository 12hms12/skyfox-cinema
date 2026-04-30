DELETE FROM predefined_avatars
WHERE url IN (
    'https://skyfoxcinema.sirv.com/Avatars/female-1.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/female-2.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/female-3.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/female-4.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/female-5.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/male-1.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/male-2.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/male-3.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/male-4.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/male-5.jpg',
    'https://skyfoxcinema.sirv.com/Avatars/neutral-1.jpg'
);

DROP TABLE IF EXISTS predefined_avatars;