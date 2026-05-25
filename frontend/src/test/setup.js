import '@testing-library/jest-dom'

// jsdom doesn't implement scrollIntoView — stub it so ChatWindow doesn't throw.
window.HTMLElement.prototype.scrollIntoView = () => {}

// jsdom doesn't implement window.confirm — default to confirming.
window.confirm = () => true
