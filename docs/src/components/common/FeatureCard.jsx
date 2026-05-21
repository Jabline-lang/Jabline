export default function FeatureCard({ icon, title, description, accent = false }) {
  return (
    <div className={`jb-card p-6 ${accent ? 'jb-card-accent' : ''}`}>
      <div className="w-10 h-10 rounded-lg bg-accent/10 border border-accent/20 flex items-center justify-center mb-4 text-xl">
        {icon}
      </div>
      <h3 className="text-jb-100 font-semibold text-base mb-2">{title}</h3>
      <p className="text-jb-400 text-sm leading-relaxed">{description}</p>
    </div>
  )
}
