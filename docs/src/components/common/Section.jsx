export default function Section({ children, title, description, id, className = '', tight = false }) {
  return (
    <section
      id={id}
      className={`${tight ? 'py-12' : 'py-20'} scroll-mt-20 ${className}`}
    >
      <div className="container-max">
        {(title || description) && (
          <div className="mb-12">
            {title && (
              <h2 className="text-jb-50 font-bold mb-4">{title}</h2>
            )}
            {description && (
              <p className="text-jb-300 text-lg max-w-2xl">{description}</p>
            )}
          </div>
        )}
        {children}
      </div>
    </section>
  )
}
