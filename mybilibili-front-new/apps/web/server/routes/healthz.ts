export default defineEventHandler(() => {
  return { status: 'ok', service: 'web', ts: Date.now() }
})