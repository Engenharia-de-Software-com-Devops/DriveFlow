type StatusTipo = 'loading' | 'success' | 'error';

interface StatusBannerProps {
  tipo: StatusTipo;
  mensagem: string;
}

export default function StatusBanner({ tipo, mensagem }: StatusBannerProps) {
  const classes = ['status-banner', `status-banner--${tipo}`];

  if (tipo === 'loading') {
    return (
      <p className={classes.join(' ')} role="status" aria-live="polite">
        {mensagem}
      </p>
    );
  }

  if (tipo === 'success') {
    return (
      <p className={classes.join(' ')} role="status" aria-live="polite">
        {mensagem}
      </p>
    );
  }

  return (
    <p className={classes.join(' ')} role="alert" aria-live="assertive">
      {mensagem}
    </p>
  );
}
