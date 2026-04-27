// place files you want to import through the `$lib` alias in this folder.
interface Repository {
    id: number; // Unique identifier for the repository
    name: string; // Repository name
    fullName: string; // Full name, including the owner (e.g., "owner/repo")
    private: boolean; // Indicates if the repository is private
    owner: {
        login: string; // Owner's username
        avatarUrl: string; // Owner's avatar URL
    };
    htmlUrl: string; // URL to the repository on GitHub
    description: string | null; // Optional repository description
    fork: boolean; // Indicates if the repository is a fork
    createdAt: string; // ISO string of when the repository was created
    updatedAt: string; // ISO string of when the repository was last updated
    pushedAt: string; // ISO string of when the repository was last pushed
    language: string | null; // Main language used in the repository
    size: number; // Size of the repository (in KB)
    visibility: "public" | "private" | "internal"; // Repository visibility
    permissions: {
        admin: boolean; // Admin access
        push: boolean; // Push access
        pull: boolean; // Pull access
    };
    linkedTo: string; // UUID of the user who linked this repository
}
