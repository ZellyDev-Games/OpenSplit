export type Category = {
    id: string;
    name: string;
};

export type GameMatch = {
    id: string;
    name: string;
    platforms: Platform[];
};

export type Platform = {
    id: string;
    name: string;
};
