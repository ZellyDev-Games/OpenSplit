export function setActiveSkin(name: string): string {
    let defaultLink = document.getElementById("default-skin") as HTMLLinkElement | null;
    let link = document.getElementById("active-skin") as HTMLLinkElement | null;
    if (!defaultLink) {
        console.log("Adding default skin link");
        defaultLink = document.createElement("link");
        defaultLink.id = "default-skin";
        defaultLink.rel = "stylesheet";
        document.head.appendChild(defaultLink);
    }
    defaultLink.href = `/skins/default/default.css`;

    if (name != "default") {
        if (!link) {
            link = document.createElement("link");
            link.id = "active-skin";
            link.rel = "stylesheet";
            document.head.appendChild(link);
        }
        link.href = `/skins/${name}/${name}.css`;
        document.documentElement.setAttribute("data-skin", name);
    }
    return name;
}
