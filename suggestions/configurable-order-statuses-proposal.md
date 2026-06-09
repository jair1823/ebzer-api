# Configurable Order Statuses - Technical Proposal

## Executive Summary

This proposal outlines a solution to simplify order management by replacing hardcoded status values with a configurable status system. The change will allow users to define custom order statuses without code changes, reducing operational complexity while maintaining system integrity through protected system statuses.

---

## Current State Analysis

### Previous Implementation
 
The system previously had **6 hardcoded statuses** defined in code:

```go
const (
    StatusConfirmed  OrderStatus = "confirmed"
    StatusInProgress OrderStatus = "in_progress"
    StatusReady      OrderStatus = "ready"
    StatusShipped    OrderStatus = "shipped"
    StatusDelivered  OrderStatus = "delivered"
    StatusCancelled  OrderStatus = "cancelled"
)
```

**Previous database constraints:**
```sql
status TEXT NOT NULL DEFAULT 'confirmed' 
CHECK(status IN ('confirmed', 'in_progress', 'ready', 'shipped', 'delivered', 'cancelled'))
```

### Problems Identified

1. **Rigidity**: Status changes require code modifications and database migrations
2. **Unnecessary Complexity**: For single-person workflows, 6 statuses may be excessive
3. **Limited Adaptability**: Cannot customize workflow based on business needs
4. **Deployment Dependencies**: Status modifications require full development cycle

---

## Proposed Solution

### Architecture Overview

Implement a two-tier status system:

#### **Tier 1: System Statuses (Protected)**
Core statuses that always exist and cannot be deleted:

- **`new`** - Initial order state
- **`completed`** - Successfully finished order
- **`cancelled`** - Cancelled order

#### **Tier 2: Custom Statuses (User-Defined)**
Configurable statuses that users can create, modify, and deactivate:

- Examples: `in_progress`, `ready_for_pickup`, `pending_payment`, `in_review`
- Each status includes:
  - **Name**: Unique identifier (slug)
  - **Display Name**: User-friendly label
  - **Color**: Hex color for UI visualization
  - **Order Position**: Determines workflow sequence
  - **Active Flag**: Enable/disable without losing history
  - **Is Final Status**: Marks order as complete

---

## Database Schema

### New Table: `order_statuses`

```sql
CREATE TABLE order_statuses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,              -- Slug identifier (e.g., 'in_progress')
    display_name TEXT NOT NULL,             -- UI display name (e.g., 'In Progress')
    color TEXT DEFAULT '#6B7280',           -- Hex color for UI
    order_position INTEGER NOT NULL,        -- Workflow sequence
    is_system_status BOOLEAN DEFAULT 0,     -- Protected system status
    is_final_status BOOLEAN DEFAULT 0,      -- Order completion flag
    is_active BOOLEAN DEFAULT 1,            -- Soft delete flag
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Indexes
CREATE INDEX idx_order_statuses_active ON order_statuses (is_active);
CREATE INDEX idx_order_statuses_order ON order_statuses (order_position);
CREATE INDEX idx_order_statuses_name ON order_statuses (name);
```

### Initial System Statuses Seed

```sql
INSERT INTO order_statuses (name, display_name, color, order_position, is_system_status, is_final_status) VALUES
('new', 'New', '#3B82F6', 1, 1, 0),
('completed', 'Completed', '#10B981', 100, 1, 1),
('cancelled', 'Cancelled', '#EF4444', 101, 1, 1);
```

### Modified Table: `orders`

```sql
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    amount_charged REAL NOT NULL,
    status_id INTEGER NOT NULL DEFAULT 1 REFERENCES order_statuses(id),
    entry_date TEXT NOT NULL DEFAULT (datetime('now')),
    estimated_delivery_date TEXT,
    delivery_type TEXT NOT NULL DEFAULT 'pickup' CHECK(delivery_type IN ('pickup', 'shipping', 'delivery')),
    client_name TEXT,
    client_phone TEXT,
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create index
CREATE INDEX idx_orders_status_id ON orders (status_id);
```

---

## Go Models

### OrderStatus Model

```go
package orders

import "creaciones-api/internal/db"

type OrderStatus struct {
    ID             int     `json:"id"`
    Name           string  `json:"name"`
    DisplayName    string  `json:"display_name"`
    Color          string  `json:"color"`
    OrderPosition  int     `json:"order_position"`
    IsSystemStatus bool    `json:"is_system_status"`
    IsFinalStatus  bool    `json:"is_final_status"`
    IsActive       bool    `json:"is_active"`
    CreatedAt      db.Time `json:"created_at"`
    UpdatedAt      db.Time `json:"updated_at"`
}

type OrderStatusCreateDTO struct {
    Name          string `json:"name" binding:"required"`
    DisplayName   string `json:"display_name" binding:"required"`
    Color         string `json:"color"`
    OrderPosition int    `json:"order_position"`
    IsFinalStatus bool   `json:"is_final_status"`
}

type OrderStatusUpdateDTO struct {
    DisplayName   *string `json:"display_name"`
    Color         *string `json:"color"`
    OrderPosition *int    `json:"order_position"`
    IsFinalStatus *bool   `json:"is_final_status"`
    IsActive      *bool   `json:"is_active"`
}

type ReorderStatusesDTO struct {
    StatusOrders []StatusOrder `json:"status_orders" binding:"required"`
}

type StatusOrder struct {
    ID       int `json:"id" binding:"required"`
    Position int `json:"position" binding:"required"`
}
```

### Updated Order Model

```go
type Order struct {
    ID                    int          `json:"id"`
    Description           string       `json:"description"`
    AmountCharged         float64      `json:"amount_charged"`
    StatusID              int          `json:"status_id"`
    Status                *OrderStatus `json:"status,omitempty"` // Populated via JOIN
    EntryDate             db.Time      `json:"entry_date"`
    EstimatedDeliveryDate *db.NullTime `json:"estimated_delivery_date"`
    DeliveryType          DeliveryType `json:"delivery_type"`
    ClientName            *string      `json:"client_name"`
    ClientPhone           *string      `json:"client_phone"`
    Notes                 *string      `json:"notes"`
    CreatedAt             db.Time      `json:"created_at"`
    UpdatedAt             db.Time      `json:"updated_at"`
}
```

---

## API Endpoints

### Order Status Management

#### List All Statuses
```
GET /api/order-statuses?active_only=true
```

**Response:**
```json
{
  "statuses": [
    {
      "id": 1,
      "name": "new",
      "display_name": "New",
      "color": "#3B82F6",
      "order_position": 1,
      "is_system_status": true,
      "is_final_status": false,
      "is_active": true
    },
    {
      "id": 4,
      "name": "in_progress",
      "display_name": "In Progress",
      "color": "#F59E0B",
      "order_position": 2,
      "is_system_status": false,
      "is_final_status": false,
      "is_active": true
    }
  ]
}
```

#### Create Custom Status
```
POST /api/order-statuses
```

**Request:**
```json
{
  "name": "ready_for_pickup",
  "display_name": "Ready for Pickup",
  "color": "#8B5CF6",
  "order_position": 3,
  "is_final_status": false
}
```

#### Update Status
```
PUT /api/order-statuses/:id
```

**Request:**
```json
{
  "display_name": "Ready to Ship",
  "color": "#10B981"
}
```

#### Deactivate Status
```
DELETE /api/order-statuses/:id
```

**Validation:**
- Cannot delete system statuses (`is_system_status = true`)
- Cannot delete statuses currently in use by orders
- Returns error with affected order count if in use

#### Reorder Statuses
```
PUT /api/order-statuses/reorder
```

**Request:**
```json
{
  "status_orders": [
    { "id": 1, "position": 1 },
    { "id": 4, "position": 2 },
    { "id": 5, "position": 3 }
  ]
}
```

### Updated Order Endpoints

Order endpoints use `status_id` as the only status input/filter:

```
GET  /api/orders?status_id=1
POST /api/orders          # Default status_id = 1 (new)
PUT  /api/orders/:id      # Accepts any active status_id
```

---

## Business Logic & Validations

### Order Status Service Validations

1. **Create Status**
   - Name must be unique (case-insensitive)
   - Name must be valid slug format (lowercase, underscores, no spaces)
   - Order position must be > 0
   - Color must be valid hex format (optional)

2. **Update Status**
   - Cannot modify `name` or `is_system_status` fields
   - Cannot set system status to inactive
   - Order position must remain unique

3. **Delete/Deactivate Status**
   - Check if status is system status → reject
   - Check if orders exist with this status → reject with count
   - Soft delete: set `is_active = false`

4. **Reorder Statuses**
   - All positions must be unique
   - System statuses can be reordered but not deleted

### Order Service Validations

1. **Create Order**
   - If no `status_id` provided, default to "new" system status
   - Validate `status_id` exists and is active

2. **Update Order**
   - Validate new `status_id` exists and is active
   - Optional: validate status transition logic (future enhancement)

---

## Migration Strategy

Because the app and database are not in active use, update the existing initial migration directly instead of creating a new migration. Any local SQLite database that already applied the old `000001` migration must be reset or recreated.

1. **Create `order_statuses` table** in `000001`
2. **Seed system statuses** (new, completed, cancelled)
3. **Seed default custom statuses** from existing hardcoded values:

```sql
INSERT INTO order_statuses (name, display_name, color, order_position, is_system_status, is_final_status) VALUES
('in_progress', 'In Progress', '#F59E0B', 2, 0, 0),
('ready', 'Ready', '#8B5CF6', 3, 0, 0),
('shipped', 'Shipped', '#06B6D4', 4, 0, 0);
```

4. **Create `orders` table** with `status_id` as the only persisted status reference
5. **Update repositories** to use `status_id` and populate status details
6. **Update services** to validate against active statuses
7. **Update handlers** to accept `status_id`
8. **Expose status management API endpoints**
9. **Update documentation and tests**

---

## Implementation Checklist

### Backend (Go)

- [ ] Create migration: `order_statuses` table
- [ ] Create initial `orders` schema with `status_id`
- [ ] Create migration: seed system statuses
- [ ] Create migration: seed default custom statuses
- [ ] Implement `OrderStatus` model
- [ ] Implement `OrderStatusRepository` (CRUD)
- [ ] Implement `OrderStatusService` (with validations)
- [ ] Implement `OrderStatusHandler` (HTTP endpoints)
- [ ] Update `OrderRepository` to use JOINs for status
- [ ] Update `OrderService` validations
- [ ] Update `OrderHandler` to accept `status_id`
- [ ] Add integration tests for status management
- [ ] Update API documentation

### Frontend (Optional - Future)

- [ ] Create Status Management UI
- [ ] Implement color picker
- [ ] Implement drag-and-drop reordering
- [ ] Update order forms to use dynamic statuses
- [ ] Update order filters with dynamic statuses
- [ ] Update order status badges with custom colors

---

## Benefits Summary

### ✅ **Flexibility**
- Add/modify statuses without code changes
- Adapt to different business workflows
- Scale from simple (3 states) to complex (10+ states)

### ✅ **Operational Simplicity**
- Users control their workflow
- No deployment needed for status changes
- Can deactivate unused statuses

### ✅ **Maintainability**
- Centralized status configuration
- Historical data preserved (soft deletes)
- Easier to extend with new features (e.g., status transitions, permissions)

### ✅ **User Experience**
- Custom colors for visual clarity
- Configurable workflow order
- Simplified UI with only relevant statuses

### ✅ **Future-Proof**
- Foundation for advanced features:
  - Status transition rules
  - Role-based status permissions
  - Status-triggered notifications
  - Custom workflows per order type

---

## Estimated Effort

| Phase | Task | Effort |
|-------|------|--------|
| 1 | Database migrations | 2 hours |
| 2 | Models & DTOs | 2 hours |
| 3 | Repository layer | 4 hours |
| 4 | Service layer + validations | 4 hours |
| 5 | Handler layer + routes | 3 hours |
| 6 | Integration tests | 3 hours |
| 7 | Update existing order endpoints | 2 hours |
| 8 | Documentation | 2 hours |
| **Total** | | **~22 hours (2.5-3 days)** |

---

## Risk Assessment

### Low Risks
- **Data Migration**: Straightforward mapping with rollback plan
- **API Compatibility**: Existing endpoints remain functional during transition

### Mitigation Strategies
1. **Reset/recreate local SQLite databases** that already applied the old `000001` migration
2. **Run migrations from a clean database** to validate the final schema
3. **Verify API clients use `status_id` only**
4. **Comprehensive testing** of status management and order flows

---

## Success Criteria

- [ ] Fresh databases are created with `status_id` as the only order status field
- [ ] Users can create/edit/delete custom statuses via API
- [ ] Orders can be filtered by custom statuses
- [ ] System statuses remain protected and functional
- [ ] API consumers use `status_id` for create/update/filter operations
- [ ] Local databases that applied the old migration are reset before testing
- [ ] Performance maintained (queries remain fast)

---

## Future Enhancements (Out of Scope)

1. **Status Transition Rules**: Define allowed transitions (e.g., `new` → `in_progress` only)
2. **Status Permissions**: Role-based access to change specific statuses
3. **Status Notifications**: Trigger events on status changes
4. **Status Templates**: Pre-configured status sets for different workflows
5. **Status Analytics**: Track time spent in each status

---

## Appendix: Example Configurations

### Configuration A: Minimal (3 states)
```
1. New ────────> 2. Completed
                     ^
                     |
                 3. Cancelled
```

### Configuration B: Standard (5 states)
```
1. New ──> 2. In Progress ──> 3. Ready ──> 4. Completed
               |
               └──────────────────────────> 5. Cancelled
```

### Configuration C: Detailed (8 states)
```
1. New ──> 2. Pending Payment ──> 3. In Progress ──> 4. Quality Check
                                        |                   |
                                        v                   v
                                   5. Ready ──────> 6. Shipped ──> 7. Completed
                                                                        ^
                                                                        |
                                                                8. Cancelled
```

---

## Questions & Answers

**Q: Can we delete the old status column immediately?**  
A: Yes. The app and database are not in active use, so the initial migration is updated directly and fresh databases are recreated without `orders.status`.

**Q: What happens to orders with deactivated statuses?**  
A: Historical orders keep their status_id. Deactivated statuses just don't appear in dropdowns.

**Q: Can we add status transition validation?**  
A: Yes, but it's a future enhancement. Initial implementation allows any status change.

**Q: How do we handle existing API integrations?**  
A: API integrations should use `status_id`. The legacy textual `status` input and `?status=` filter are intentionally removed.

---

## Conclusion

This proposal provides a flexible, maintainable solution for order status management that eliminates hardcoded limitations while preserving system integrity through protected system statuses. Because the app and database are not in active use, the initial schema can be updated directly and validated from a clean database.

**Recommended Next Steps:**
1. Review and approve proposal
2. Create implementation branch
3. Update existing migration and backend files
4. Reset local database and run tests

---

**Document Version:** 1.0  
**Date:** 2026-06-08  
**Status:** Pending Approval
