import type { AuthRecord } from 'pocketbase'
import { writable } from 'svelte/store'

export const currentUser = writable<AuthRecord | null>()

export const welcomeMessage = writable<boolean>(false)