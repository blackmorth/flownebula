import AdminLayout from "../components/AdminLayout";
import AdminUsersView from "../views/AdminUsersView";
import { useEffect, useState } from "react";
import { api } from "../services/api";

export default function AdminUsers() {
    const [loading, setLoading] = useState(true);
    const [users, setUsers] = useState([]);

    const token = localStorage.getItem("token");

    const loadUsers = () => {
        api("GET", "/admin/users", null, token).then((res) => {
            setUsers(Array.isArray(res) ? res : []);
            setLoading(false);
        });
    };

    useEffect(() => {
        loadUsers();
    }, []);

    const enableAgent = (id) =>
        api("POST", `/admin/users/${id}/agent/enable`, null, token).then(loadUsers);

    const disableAgent = (id) =>
        api("POST", `/admin/users/${id}/agent/disable`, null, token).then(loadUsers);

    const regenerateToken = (id) =>
        api("POST", `/admin/users/${id}/agent/regenerate`, null, token).then(loadUsers);

    return (
        <AdminLayout>
            <AdminUsersView
                loading={loading}
                users={users}
                enableAgent={enableAgent}
                disableAgent={disableAgent}
                regenerateToken={regenerateToken}
            />
        </AdminLayout>
    );
}
