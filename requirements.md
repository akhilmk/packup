* LOGIN
- admin and user login

* ADMIN SIDE
- admin page contain Users tab, Default Task tab, and My Todos page

 ** Users tab
- admin can see all the users in his homepage's users tab,
- admin can click on user's 'View Todos', admin can see that users todo status and its progress.
- admin can add new custom task under a user, this is shared by default and admin can hide also for admin only visibility
- admin added task can have label like 'admin'
- admin added custom task under user will be visible to user, there will be label 'shared' along with 'admin' label
- admin added a custom task under a user, have a option to edit, delete , change progress - for ADMIN
- admin added tasks, customer can only update progress of shared task.
** Labels & Visibility Rules
 - Three distinct type labels: 'default', 'admin', 'user'
 - One status label: 'shared'
 - 'Shared' label logic: Shows only when the viewer is NOT the creator of the todo item, AND the item is not 'admin' created (since admin implies shared).
   - Example 1: User viewing Admin-created task -> Shows 'admin' (No 'shared' label)
   - Example 2: Admin viewing User-created task -> Shows 'user' + 'shared'
   - Example 3: User viewing User-created task -> Shows 'user' (No 'shared' label)
   - Example 4: Admin viewing Admin-created task (assigned to user) -> Shows 'admin'
### Custom Tasks (User Specific)
*   **Concept**:
    *   **User Created**: Users create their own tasks. By default, these are "Personal". Users can choose to "Share with Admin".
    *   **Admin Created**: Admins can create tasks *for* a specific user. These are automatically "Shared".
*   **Rules**:
    *   **Status (Progress)**: **Shared Responsibility**. Both the Admin and the User can update the status (Pending <-> In Progress <-> Done) of *any* shared task, regardless of who created it. This reflects a collaborative effort.
    *   **Edit (Text) / Delete**: **Creator Only**.
        *   If **User** created it: Only User can edit text or delete. Admin is Read-Only for text/existence.
        *   If **Admin** created it: Only Admin can edit text or delete. User cannot delete (but might be able to hide it, see below).
    *   **Visibility**:
        *   **Admin View**:
            *   User-created (Shared): Labeled "User" + "Shared". Status: Editable. Text: Read-only.
            *   Admin-created: Labeled "Shared" (if visible to user) or "Hidden from User" (if hidden). Status: Editable. Text: Editable.
            *   **Hidden Tasks**: If an Admin-created task is hidden from the user, it is labeled "hidden from user".
        *   **User View**:
            *   User-created: Standard view. If hidden from admin (unshared), labeled "hidden from admin".
            *   Admin-created: Labeled "Admin". Status: Editable. Text: Read-only.
    *   **Hiding**:
        *    Admin can "Hide" an Admin-created task from the User (e.g., drafting).
        *    User can "Hide" their task from Admin (unshare).
 
 USER/CUSTOMER SIDE
 - users login show all 'default' labled tasks, 'admin' labled tasks.
 - user can make only progress change for 'default' and 'admin' task.
 - user can add new task under him, lablel will be 'user', he can edit, delete, change progress for his task.
 - user can hide his task (unshare). If hidden (unshared), User sees 'hidden from admin'. If shared, Admin sees 'user' + 'shared'.
 - admin added custom todos display as 'admin' in user view.
   - Both User and Admin can update the progress of these tasks.
 - all 'user' created task's deletion/edit permitted only for user. Admin can only view (and see 'shared').
