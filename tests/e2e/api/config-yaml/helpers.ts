export * from '../../helpers.js';
// config-yaml feature uses apiUrl (alias for apiBaseUrl)
export const apiUrl = process.env.E2E_API_URL ?? 'http://localhost:8080';
