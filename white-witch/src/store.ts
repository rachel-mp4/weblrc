import { writable } from "svelte/store"

export const messages = writable<Message[]>([]);

export const topic = writable("loading...")

export type Message = {
    id: number
    color: number
    name: string
    text: string
    active: boolean
}