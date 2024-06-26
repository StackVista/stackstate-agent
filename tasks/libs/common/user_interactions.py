from .color import color_message


def yes_no_question(input_message, color="white", default=None):
    choice = None
    valid_answers = {'yes': True, 'y': True, 'no': False, 'n': False, '': default}

    if default is None:
        default_answer_prompt = "[y/n]"
    elif default:
        default_answer_prompt = "[Y/n]"
    else:
        default_answer_prompt = "[y/N]"

    while choice not in valid_answers or valid_answers[choice] is None:
        print(color_message("{} {} ".format(input_message, default_answer_prompt), color))#, end='')  # sts - 'end' param unsupported by py2
        choice = input().strip().lower()

    return valid_answers[choice]
