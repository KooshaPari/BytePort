interface Owner {
    login: string;
    id: number;
    node_id: string;
    avatar_url: string;
    html_url: string;
    type: string;
    site_admin: boolean;
}

interface Permissions {
    admin: boolean;
    maintain: boolean;
    push: boolean;
    triage: boolean;
    pull: boolean;
}

interface Repository {
    id: number;
    node_id: string;
    name: string;
    full_name: string;
    private: boolean;
    owner: Owner;
    html_url: string;
    description: string | null;
    fork: boolean;
    created_at: string;
    updated_at: string;
    pushed_at: string;
    git_url:
