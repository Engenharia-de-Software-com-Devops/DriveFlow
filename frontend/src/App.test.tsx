// src/App.test.tsx
import { render, screen } from '@testing-library/react';
import App from './App';

test('renderiza o conteúdo principal', () => {
  render(<App />);
    expect(screen.getByText(/get started/i)).toBeInTheDocument();
});