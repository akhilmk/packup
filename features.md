* LOGIN
- admin and user login

* ADMIN SIDE
- admin page contain Users tab, Default Task tab, and My Todos page
 ** Users tab
 - admin can see all users in his homepage users tab, by clicking each user -> admin can see each users todo status and its progress.
 - admin can add new custom task under a user, this is shared by default and admin can hide also as admin visibility
  - if admin added a custom task under a user, have a option to edit, delete this task for admin only, other default tasks can be edit deleted from default task page
  - all adminshared task delete, edit permission will be only for admin, customer can only update progress of shared task
- shared task can have label like admin shared or something 
--  admin shared todo with user, with label 'custom task' with all permission for admin, user can only update progress


** Default Task tab
- admin can add default tasks, this will go list under users with label 'default task'
- admin can edit, update, delete this task from Default task menu only.

** Admin My Todos 
- admin will have its own todo page to manage his todo items , unrelated to any user
- he can add, edit, delete task, default task should not come here

USER/CUSTOMER SIDE
-users login show all default task todo items added by admin
- users can edit progress of custom task no other oprion
- user can add new todo for him, by default admin can see this through admin users list-> user todo way, user can hide it if needed by checkbox, user created and shared label can be shown
  - all user created and shared task delete edit change progress permission will be only for user, admin can only see no edit
  - admin shared todo task display in user as admin created label, but do not show the option to hide for user