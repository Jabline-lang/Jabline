/**
 * Utilidades de formato y manipulación de strings
 */

export const formatCode = (code) => {
  return code.trim()
}

export const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch (err) {
    console.error('Failed to copy:', err)
    return false
  }
}

export const slugify = (str) => {
  return str
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^\w-]/g, '')
}

export const truncate = (str, length) => {
  return str.length > length ? str.substring(0, length) + '...' : str
}
