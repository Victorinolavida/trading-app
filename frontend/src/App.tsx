import { Sidebar } from "./components/Sidebar";
import { Outlet } from "react-router";

// type AuthView = "login" | "register";

// function getSecurityIdFromHash(hash: string): string | null {
//   if (!hash) return null;
//   // support #/securities/:id or #securities/:id
//   const m = hash.match(/securities[/:](.+)$/);
//   if (m && m[1]) {
//     // strip trailing slashes
//     return decodeURIComponent(m[1].replace(/\/+$/, ""));
//   }
//   return null;
// }

// function getBlotterFromHash(hash: string): boolean {
//   if (!hash) return false;
//   return hash === "#/blotter" || hash === "#/blotter/";
// }

export default function App() {
  //   const [view, setView] = useState<AuthView>("login");
  //   const { data: user, isLoading, isError } = useSession();
  //   const [hash, setHash] = useState(window.location.hash);

  //   useEffect(() => {
  //     const onHash = () => setHash(window.location.hash);
  //     window.addEventListener("hashchange", onHash);
  //     return () => window.removeEventListener("hashchange", onHash);
  //   }, []);

  //   const securityId = getSecurityIdFromHash(hash);

  //   if (isLoading) {
  //     return <div className="flex min-h-screen items-center justify-center p-4">Loading session...</div>;
  //   }

  //   if (isError) {
  //     return <div className="flex min-h-screen items-center justify-center p-4 text-sm text-red-400">Could not reach the gateway.</div>;
  //   }

  //   if (!user) {
  //     return (
  //       <div className="flex min-h-screen items-center justify-center p-4">
  //         {view === "login" ? (
  //           <LoginPage onSwitch={() => setView("register")} />
  //         ) : (
  //           <RegisterPage onSwitch={() => setView("login")} />
  //         )}
  //       </div>
  //     );
  //   }

  //   if (securityId) {
  //     return (
  //       <div className="min-h-screen overflow-x-hidden">
  //         <SecurityDetailPage securityId={securityId} />
  //       </div>
  //     );
  //   }

  //   if (getBlotterFromHash(hash)) {
  //     return <BlotterPage />;
  //   }

  return (
    <div className="min-h-screen overflow-x-hidden">
      <div className="flex min-h-screen overflow-x-hidden">
        <Sidebar />
        <main className="ml-64 flex-1 min-h-screen overflow-x-hidden p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
