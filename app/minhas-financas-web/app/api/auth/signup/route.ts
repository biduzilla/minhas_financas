import { signup } from "@/lib/api/endpoints/auth";
import { SignUpInput } from "@/types/api";

export async function POST(req: Request) {
    let input: SignUpInput;
    try {
        input = await req.json();
    } catch {
        return Response.json({ message: 'invalid JSON body' }, { status: 400 });
    }

    const result = await signup(input);

    if (!result.ok) {
        return Response.json(result.error.body, { status: result.error.status });
    }

    return Response.json(result.data, { status: 201 });
}