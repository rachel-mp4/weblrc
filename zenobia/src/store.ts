import { signal } from "@preact/signals";

export const messages = signal<message[]>([])

export type message = {
    id: number
    color: number
    name: string
    text: string
    active: boolean
}

export const topic = signal("loading...")