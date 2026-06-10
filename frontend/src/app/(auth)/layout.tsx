export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[hsl(210,40%,98%)] px-4 py-12">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="w-10 h-10 mx-auto mb-3 rounded-xl bg-primary flex items-center justify-center">
            <span className="text-white font-bold text-sm">SF</span>
          </div>
          <h1 className="text-xl font-bold text-gray-900 tracking-tight">StayFlow Track</h1>
          <p className="text-sm text-muted-foreground mt-1">Service Apartment Management</p>
        </div>
        {children}
      </div>
    </div>
  );
}
