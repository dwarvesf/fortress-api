INSERT INTO public.discord_log_templates (id, deleted_at, created_at, updated_at, type, content) VALUES
('d511ac81-e098-44a3-8712-3e0cbfa57386', null, '2023-05-09 06:25:33.019333', '2023-05-09 06:25:33.019333', 'employee_update_base_salary', '{{ employee_id }} update {{ updated_employee_id }} salary to new one: {{ new_salary }}.'),
('3fc86aaf-32ed-4c44-a100-69c6d1c43db5', null, '2023-05-09 06:25:18.021158', '2023-05-09 06:25:18.021158', 'employee_update_working_status', '{{ employee_id }} update {{ updated_employee_id }} working status to {{ working_status }}.'),
('a1de8a0a-4e91-45c4-b093-5f525b1adbda', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'project_member_add', '{{ employee_id }} add {{ updated_employee_id }} to project {{ project_name }} as {{ deployment_type }} deployment'),
('3ced7070-959f-408b-9155-00cdab019fab', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'project_member_remove', '{{ employee_id }} remove {{ updated_employee_id }} to project {{ project_name }}'),
('f6ed892e-5c3c-4bca-a738-9805e767084c', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'invoice_send', '{{ employee_id }} just sent invoice #{{ invoice_number }}.'),
('15bc0057-42df-4bca-9749-9e1a3a07970d', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'invoice_paid', 'Invoice #{{ invoice_number }} paid.'),
('9616d2bd-22d7-43b4-8a2a-067927b0d875', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'project_member_update_status', '{{ employee_id }} update {{ updated_employee_id }} project member status in {{ project_name }} to {{ status }}.'),
('a427c903-2932-408c-8f55-f7e0213a1432', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'project_member_update_charge_rate', '{{ employee_id }} update {{ updated_employee_id }} charge rate in {{ project_name }} to {{ rate }}.'),
('4a0c264a-0b90-48d7-b217-19d9a5d66bc7', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'payroll_commit', 'payroll batch.{{ batch_number }}-{{ month }}-{{ year }} is committed.'),
('b15fd037-ffac-407d-b878-10809236868f', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'employee_submit_onboarding_form', '{{ employee_id }} submit new employee onboarding form.  '),
('cb2610ac-5c7d-45f7-8d71-50c815526a75', null, '2023-05-09 06:26:16.634402', '2023-05-09 06:26:16.634402', 'project_member_update_end_date', '{{ employee_id }} update {{ updated_employee_id }} end date in {{ project_name }} to {{ end_date }}.');