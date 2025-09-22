export default class StatTime {
    raw: number = 0;
    formatted: string = "00:00:00.00"

    constructor(raw: number = 0, formatted: string = "00:00:00.00") {
        this.raw = raw;
        this.formatted = formatted;
    }
}
