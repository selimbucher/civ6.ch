-- Renames
UPDATE achievements SET name = 'Smurf',
    description = 'Enter a game with the lowest rating in the lobby and finish with the highest rating overall.'
    WHERE id = 15;
UPDATE achievements SET name = 'Roulette'          WHERE id = 11;
UPDATE achievements SET name = 'Reveros Paragon'   WHERE id = 9;

-- Move active rare (diff=2) → uncommon (diff=1), keep No Mither (16) and The Answer to Everything (18) as rare
UPDATE achievements SET difficulty = 1
    WHERE difficulty = 2 AND disabled = false AND id NOT IN (16, 18);

-- New rare achievements
INSERT INTO achievements (name, description, difficulty, game, stats, exclusive, disabled) VALUES
    ('Unpolitical',
     'End a game with at least 2000 diplomatic favor.',
     2, '{"favor": {"min": 2000}}', '{}', false, false),

    ('Fairy of Schorenhausen',
     'End a game with at least 2000 science and 2000 culture per turn.',
     2, '{"science": {"min": 2000}, "culture": {"min": 2000}}', '{}', false, false),

    ('Farming Simulator',
     'Achieve a winning streak of at least 10.',
     2, '{}', '{"winning_streak": {"min": 10}}', false, false),

    ('El Salvador',
     'Achieve a non-capitulation victory without ever researching Mining.',
     3, '{"victorious": true, "victory_type": {"not": true, "value": "Capitulation"}, "not_mining": true}',
     '{}', false, false);

-- New common achievements — one per victory type (except Score and Territorial)
INSERT INTO achievements (name, description, difficulty, game, stats, exclusive, disabled) VALUES
    ('Conqueror',    'Win a Domination victory.',   0, '{"victorious": true, "victory_type": "Domination"}',  '{}', false, false),
    ('Crusader',     'Win a Religious victory.',    0, '{"victorious": true, "victory_type": "Religious"}',   '{}', false, false),
    ('Tech Pioneer', 'Win a Science victory.',      0, '{"victorious": true, "victory_type": "Science"}',     '{}', false, false),
    ('Renaissance',  'Win a Culture victory.',      0, '{"victorious": true, "victory_type": "Culture"}',     '{}', false, false),
    ('Statesman',    'Win a Diplomatic victory.',   0, '{"victorious": true, "victory_type": "Diplomatic"}',  '{}', false, false),
    ('Ruthless',     'Win a Capitulation victory.', 0, '{"victorious": true, "victory_type": "Capitulation"}','{}', false, false);
