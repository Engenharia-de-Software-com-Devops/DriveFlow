interface Props {
  onClick: () => void;
  disabled?: boolean;
}

export default function BotaoPreenchimentoAleatorio({ onClick, disabled }: Props) {
  return (
    <button
      type="button"
      className="botao-secundario"
      onClick={onClick}
      disabled={disabled}
      aria-label="Preenchimento aleatório"
    >
      Preenchimento aleatório
    </button>
  );
}
