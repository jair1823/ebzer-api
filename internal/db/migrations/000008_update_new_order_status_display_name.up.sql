UPDATE order_statuses
SET display_name = 'Pendiente',
    updated_at = datetime('now')
WHERE name = 'new';
