// web/components/auth/AuthDivider.tsx
export default function AuthDivider() {
  return (
    <div className="flex items-center gap-3 my-2">
      <div className="flex-1 h-px bg-white/10" />
      <span className="text-xs text-gray-500 select-none">or</span>
      <div className="flex-1 h-px bg-white/10" />
    </div>
  )
}
