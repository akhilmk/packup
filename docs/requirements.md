# PackUp Requirements & Functionality

PackUp is a collaborative task management application designed for streamlined interaction between administrators and customers.

## 1. Authentication & Roles

- **Google OAuth**: Users sign in securely using their Google accounts.
- **Role Detection**:
  - **Admin**: Users whose emails are listed in the `ADMIN_EMAILS` environment variable.
  - **Customer**: All other authenticated users.

---

## 2. Admin Experience

The Admin Dashboard is the central hub for managing global tasks and individual user progress.

### A. Users Tab (User Management)
- **User Directory**: Lists all registered users with their name, email, role, and join date.
- **Progress Tracking**: Clicking **"View Todos"** on any user opens their personal task environment for management.

### B. Default Tasks Tab (Global Management)
- **Concept**: Tasks created here are automatically assigned to **every user** in the system.
- **Operations**: Add, Edit, and Delete global tasks.
- **Individual Tracking**: While the task text is global, each user tracks their own progress (Pending, In Progress, Done) independently.

### C. My Todos View (Personal Admin Tasks)
- **Concept**: Admins can toggle between the Dashboard and their personal task list.
- **Personal Workspace**: This view focuses on the admin's own tasks, excluding managed user tasks for better focus.

### D. Managing User-Specific Todos
When managing a specific user, the admin sees three categorized sections:
1.  **Default**: Global tasks currently assigned to the user.
2.  **Admin Added**: Tasks specifically created by the admin for this user.
    - **Visibility**: Can be "Shared" (visible to user) or "Hidden from User".
3.  **Customer Added**: Tasks created by the user and shared with the admin.

**Admin Capabilities:**
- **Create**: Add custom tasks for the user with an optional "Hide from user" flag.
- **Modify**: Edit text, toggle visibility, or delete tasks created by an admin.
- **Collaborate**: Update the status (Pending <-> In Progress <-> Done) of **any** task visible in the user's dashboard.

---

## 3. Customer (User) Experience

Customers see a clean, sectioned view of all tasks assigned to them.

### A. Task Sources
1.  **Default**: Global system requirements (Labeled: **Default Task**).
2.  **Admin Added**: Tasks assigned specifically by an administrator (Labeled: **Admin**).
3.  **Personal Tasks**: Tasks created by the user (Labeled: **Shared** or **Hidden from Admin**).

### B. Personal Task Management
- **Create**: Add personal tasks.
- **Sharing**: Toggle **"Share with Admin"** to allow administrators to see and update the status of personal tasks.
- **Control**: Edit text (double-click) or delete personal tasks.

### C. Task Interaction
- **Status Updates**: Users can update the progress of **any** task in their list.
- **Read-Only Text**: Users cannot edit the text or delete tasks sourced from the Admin or Default lists.

---

## 4. UI Labels & Permissions

| Label | Logic | Permissions (Text/Delete) |
| :--- | :--- | :--- |
| <span style="color: #9333ea">**Default Task**</span> | Assigned to all users globally | Admin Only |
| <span style="color: #2563eb">**Admin**</span> | Assigned to a user by an admin | Admin Only |
| <span style="color: #16a34a">**Shared**</span> | User-created and visible to Admin | User Only |
| <span style="color: #64748b">**Hidden from Admin**</span> | User-created and private | User Only |
| <span style="color: #64748b">**Hidden from User**</span>| Admin-created and private | Admin Only |

### Shared Responsibility Logic
- **Progress Tracking**: For all visible tasks, both the Admin and the User can update the status (e.g., marking as "Done").
- **Content Integrity**: Only the original creator (Admin or User) can edit the task text or delete it.
